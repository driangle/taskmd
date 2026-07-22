#!/usr/bin/env python3
"""Execute an approved safe-queue proposal with mutation ordering guarantees."""

from __future__ import annotations

import argparse
import contextlib
import json
import os
import shutil
import subprocess
import sys
from pathlib import Path

import safe_queue


def command(args: list[str], cwd: Path) -> subprocess.CompletedProcess[str]:
    return subprocess.run(args, cwd=cwd, text=True, capture_output=True, check=False)


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser()
    parser.add_argument("task", nargs="?")
    parser.add_argument("--project-root", default=".")
    parser.add_argument("--main")
    parser.add_argument("--allow-stacked", action="store_true")
    parser.add_argument("--approved", action="store_true")
    parser.add_argument("--automatic", action="store_true")
    parser.add_argument("--expected-decision")
    parser.add_argument("--expected-base-branch")
    parser.add_argument("--expected-base-commit")
    parser.add_argument("--expected-approval-token")
    parser.add_argument("--owner")
    parser.add_argument("--state-file", help=argparse.SUPPRESS)
    return parser.parse_args()


def locate_created_worktree(root: Path, branch: str) -> Path | None:
    output = command(["git", "worktree", "list", "--porcelain"], root)
    if output.returncode:
        return None
    current_path: Path | None = None
    for line in output.stdout.splitlines():
        if line.startswith("worktree "):
            current_path = Path(line.removeprefix("worktree "))
        elif line == f"branch refs/heads/{branch}":
            return current_path
    return None


def untracked_task_path(root: Path, file_path: str) -> Path | None:
    source = Path(file_path)
    if not source.is_absolute():
        source = root / source
    if not source.exists():
        return None
    try:
        relative = source.resolve().relative_to(root.resolve())
    except ValueError as error:
        raise safe_queue.CollectionError("task file is outside the project root") from error
    tracked = command(["git", "ls-files", "--error-unmatch", str(relative)], root)
    if tracked.returncode == 0:
        return None
    return relative


def copy_task(root: Path, destination: Path, relative: Path) -> None:
    source = root / relative
    target = destination / relative
    target.parent.mkdir(parents=True, exist_ok=True)
    shutil.copy2(source, target)


def fail(reason: str, assessment: dict | None = None) -> int:
    json.dump(
        {"executed": False, "reason": reason, "assessment": assessment},
        sys.stdout,
        sort_keys=True,
    )
    sys.stdout.write("\n")
    return 1


def approval_matches(args: argparse.Namespace, assessment: dict) -> str | None:
    if not args.approved:
        return None
    required = {
        "--expected-decision": args.expected_decision,
        "--expected-base-branch": args.expected_base_branch,
        "--expected-base-commit": args.expected_base_commit,
        "--expected-approval-token": args.expected_approval_token,
    }
    missing = [flag for flag, value in required.items() if not value]
    if missing:
        return f"approved execution requires {' '.join(missing)}"
    proposal = assessment["proposal"]
    expected = (
        args.expected_decision,
        args.expected_base_branch,
        args.expected_base_commit,
    )
    actual = (
        assessment["decision"],
        proposal["base_branch"],
        proposal["base_commit"],
    )
    if expected != actual:
        return "fresh assessment no longer matches the explicitly approved proposal"
    if args.expected_approval_token != assessment.get("approval_token"):
        return "fresh assessment no longer matches the full explicitly approved proposal"
    return None


@contextlib.contextmanager
def execution_lock(root: Path):
    common_dir = safe_queue.git(root, "rev-parse", "--git-common-dir", check=False)
    lock_parent = Path(common_dir) if common_dir else root
    if not lock_parent.is_absolute():
        lock_parent = root / lock_parent
    lock_path = lock_parent / "taskmd-safe-queue.lock"
    try:
        lock_path.mkdir()
    except FileExistsError as error:
        raise safe_queue.CollectionError(
            "another safe-queue execution is already in progress"
        ) from error
    try:
        (lock_path / "owner").write_text(f"pid={os.getpid()}\n")
        yield
    finally:
        owner = lock_path / "owner"
        if owner.exists():
            owner.unlink()
        lock_path.rmdir()


def execute_locked(args: argparse.Namespace, root: Path) -> int:
    try:
        if args.state_file:
            state = json.loads(Path(args.state_file).read_text())
        elif args.task:
            state = safe_queue.collect_state(root, args.task, args.main)
        else:
            return fail("task is required unless --state-file is used")
        assessment = safe_queue.assess(state, allow_stacked=args.allow_stacked)
    except (OSError, KeyError, json.JSONDecodeError, safe_queue.CollectionError) as error:
        return fail(str(error))

    decision = assessment["decision"]
    if decision not in {"START_FROM_MAIN", "START_STACKED"}:
        return fail("assessment does not permit execution", assessment)
    if args.automatic and (
        decision != "START_FROM_MAIN" or not assessment.get("automatic_allowed")
    ):
        return fail("automatic mode is restricted to unambiguous START_FROM_MAIN", assessment)
    if decision == "START_STACKED" and not args.approved:
        return fail("stacked execution requires explicit approval", assessment)
    approval_error = approval_matches(args, assessment)
    if approval_error:
        return fail(approval_error, assessment)

    proposal = assessment["proposal"]
    branch = proposal["worktree_name"]
    try:
        untracked_task = untracked_task_path(root, assessment["task"]["file_path"])
    except safe_queue.CollectionError as error:
        return fail(str(error), assessment)
    create = command(
        [
            "wt",
            "switch",
            "--create",
            branch,
            "--base",
            proposal["base_commit"],
            "--no-cd",
            "--format",
            "json",
        ],
        root,
    )
    if create.returncode != 0:
        return fail(
            f"worktree creation failed: {create.stderr.strip() or create.stdout.strip()}",
            assessment,
        )

    destination = locate_created_worktree(root, branch)
    if destination is None:
        return fail("worktree was created but its path could not be identified", assessment)

    metadata = command(
        ["git", "config", f"branch.{branch}.taskmd-task-id", assessment["task"]["id"]],
        root,
    )
    if metadata.returncode != 0:
        return fail("worktree created but task-to-branch metadata could not be recorded", assessment)
    base_metadata = {
        "taskmd-base-branch": proposal["base_branch"],
        "taskmd-base-commit": proposal["base_commit"],
    }
    for key, value in base_metadata.items():
        metadata = command(["git", "config", f"branch.{branch}.{key}", value], root)
        if metadata.returncode != 0:
            return fail("worktree created but exact base metadata could not be recorded", assessment)

    set_args = ["taskmd", "set", assessment["task"]["id"], "--status", "in-progress"]
    if args.owner:
        set_args.extend(["--owner", args.owner])
    claim_cwd = root if untracked_task else destination
    claimed = command(set_args, claim_cwd)
    if claimed.returncode != 0:
        return fail(
            f"worktree created but task claim failed: {claimed.stderr.strip() or claimed.stdout.strip()}",
            assessment,
        )
    if untracked_task:
        copy_task(root, destination, untracked_task)

    json.dump(
        {
            "executed": True,
            "decision": decision,
            "branch": branch,
            "worktree": str(destination),
            "base_branch": proposal["base_branch"],
            "base_commit": proposal["base_commit"],
            "merge_order": proposal["merge_order"],
            "post_merge": proposal["post_merge"],
        },
        sys.stdout,
        sort_keys=True,
    )
    sys.stdout.write("\n")
    return 0


def main() -> int:
    args = parse_args()
    root = Path(args.project_root).resolve()
    if not args.approved and not args.automatic:
        return fail("execution requires --approved or explicitly selected --automatic")
    if args.state_file and os.environ.get("TASKMD_SAFE_QUEUE_TESTING") != "1":
        return fail("--state-file is restricted to the test harness")
    try:
        with execution_lock(root):
            return execute_locked(args, root)
    except (OSError, safe_queue.CollectionError) as error:
        return fail(str(error))


if __name__ == "__main__":
    raise SystemExit(main())
