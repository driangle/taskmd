#!/usr/bin/env python3
"""Deterministically assess whether a task can start in another worktree."""

from __future__ import annotations

import argparse
import hashlib
import json
import re
import subprocess
import sys
from pathlib import Path
from typing import Any

DECISIONS = {"START_FROM_MAIN", "START_STACKED", "WAIT", "UNSAFE"}
SAFE_TASK_ID = re.compile(r"^[A-Za-z0-9][A-Za-z0-9._-]*$")


class CollectionError(RuntimeError):
    """Raised when required assessment evidence cannot be collected."""


def run(command: list[str], cwd: Path, check: bool = True) -> str:
    result = subprocess.run(
        command, cwd=cwd, text=True, capture_output=True, check=False
    )
    if check and result.returncode != 0:
        detail = result.stderr.strip() or result.stdout.strip()
        raise CollectionError(f"{' '.join(command)} failed: {detail}")
    return result.stdout


def json_command(command: list[str], cwd: Path) -> Any:
    output = run(command, cwd)
    try:
        return json.loads(output)
    except json.JSONDecodeError as error:
        raise CollectionError(f"{' '.join(command)} returned invalid JSON") from error


def git(cwd: Path, *args: str, check: bool = True) -> str:
    return run(["git", "-C", str(cwd), *args], cwd, check=check).strip()


def default_branch(root: Path) -> str:
    remote_head = git(
        root, "symbolic-ref", "--quiet", "--short", "refs/remotes/origin/HEAD", check=False
    )
    if remote_head:
        return remote_head.removeprefix("origin/")
    for candidate in ("main", "master"):
        if git(root, "show-ref", "--verify", f"refs/heads/{candidate}", check=False):
            return candidate
    raise CollectionError("could not determine the main branch")


def worktree_records(root: Path, main: str) -> list[dict[str, Any]]:
    output = git(root, "worktree", "list", "--porcelain")
    records: list[dict[str, str]] = []
    current: dict[str, str] = {}
    for line in output.splitlines() + [""]:
        if not line:
            if current:
                records.append(current)
                current = {}
            continue
        key, _, value = line.partition(" ")
        current[key] = value

    result: list[dict[str, Any]] = []
    for record in records:
        path = Path(record["worktree"])
        branch = record.get("branch", "").removeprefix("refs/heads/")
        branch_files, working_files = changed_files(path, main)
        head = record.get("HEAD", "")
        ahead = git(path, "rev-list", "--count", f"{main}..HEAD", check=False)
        result.append(
            {
                "path": str(path),
                "branch": branch,
                "head": head,
                "clean": not working_files,
                "commits_ahead": int(ahead or "0"),
                "head_merged_to_main": bool(
                    head
                    and subprocess.run(
                        ["git", "-C", str(path), "merge-base", "--is-ancestor", head, main],
                        capture_output=True,
                        check=False,
                    ).returncode
                    == 0
                ),
                "branch_changed_files": branch_files,
                "working_changed_files": working_files,
                "changed_files": sorted(set(branch_files) | set(working_files)),
                "task_id": branch_config(root, branch, "taskmd-task-id"),
                "frozen": branch_config(root, branch, "taskmd-frozen") == "true",
                "validated_commit": branch_config(
                    root, branch, "taskmd-validated-commit"
                ),
            }
        )
    return result


def branch_records(root: Path, main: str) -> list[dict[str, Any]]:
    output = git(
        root,
        "for-each-ref",
        "--format=%(refname:short)%00%(objectname)",
        "refs/heads",
    )
    records = []
    for line in output.splitlines():
        branch, _, head = line.partition("\x00")
        task_id = branch_config(root, branch, "taskmd-task-id")
        if not task_id:
            continue
        records.append(
            {
                "branch": branch,
                "head": head,
                "task_id": task_id,
                "head_merged_to_main": commit_is_ancestor(root, head, main),
                "validated_commit": branch_config(
                    root, branch, "taskmd-validated-commit"
                ),
            }
        )
    return records


def commit_is_ancestor(root: Path, commit: str, descendant: str) -> bool:
    if not commit:
        return False
    return (
        subprocess.run(
            ["git", "-C", str(root), "merge-base", "--is-ancestor", commit, descendant],
            capture_output=True,
            check=False,
        ).returncode
        == 0
    )


def changed_files(path: Path, main: str) -> tuple[list[str], list[str]]:
    branch_names = set(git(path, "diff", "--name-only", f"{main}...HEAD", check=False).splitlines())
    working_names: set[str] = set()
    commands = [
        ["diff", "--name-only"],
        ["diff", "--cached", "--name-only"],
        ["ls-files", "--others", "--exclude-standard"],
    ]
    for args in commands:
        output = git(path, *args, check=False)
        working_names.update(line for line in output.splitlines() if line)
    return sorted(branch_names - {""}), sorted(working_names)


def branch_config(root: Path, branch: str, key: str) -> str:
    if not branch:
        return ""
    return git(root, "config", "--get", f"branch.{branch}.{key}", check=False)


def literal_touch_path(value: str) -> bool:
    if not value or any(character.isspace() for character in value):
        return False
    path = Path(value)
    if path.is_absolute():
        return False
    return "/" in value or value.startswith(".") or bool(path.suffix)


def context_paths(root: Path, task: dict[str, Any]) -> dict[str, Any]:
    task_id = task["id"]
    context = json_command(
        ["taskmd", "context", "--task-id", task_id, "--format", "json"], root
    )
    files = context.get("files") or []
    scope_files = [
        item for item in files if item.get("source", "").startswith("scope:")
    ]
    resolved_scopes = {
        item["source"].removeprefix("scope:")
        for item in scope_files
        if item.get("source")
    }
    literal_paths: set[str] = set()
    unresolved: set[str] = set()
    for touch in task.get("touches") or []:
        if touch in resolved_scopes:
            continue
        if literal_touch_path(touch):
            literal_paths.add(touch)
        else:
            unresolved.add(touch)
    return {
        "paths": sorted(
            literal_paths
            | {item["path"] for item in scope_files if item.get("path")}
        ),
        "missing_scope": sorted(
            {
                item["path"]
                for item in scope_files
                if item.get("path") and not item.get("exists")
            }
        ),
        "missing_context": sorted(
            {
                item["path"]
                for item in files
                if item.get("source") == "explicit"
                and item.get("path")
                and not item.get("exists")
            }
        ),
        "unresolved": sorted(unresolved),
    }


def frontmatter_list(path: str, field: str) -> list[str] | None:
    try:
        lines = Path(path).read_text().splitlines()
    except OSError:
        return None
    if not lines or lines[0].strip() != "---":
        return None
    inline = re.compile(rf"^{re.escape(field)}:\s*\[(.*)]\s*$")
    block = re.compile(rf"^{re.escape(field)}:\s*$")
    for index, line in enumerate(lines[1:], start=1):
        if line.strip() == "---":
            break
        match = inline.match(line)
        if match:
            return [
                item.strip().strip("\"'")
                for item in match.group(1).split(",")
                if item.strip()
            ]
        if block.match(line):
            values: list[str] = []
            for child in lines[index + 1 :]:
                if not child.startswith((" ", "\t")):
                    break
                stripped = child.strip()
                if stripped.startswith("- "):
                    values.append(stripped[2:].strip().strip("\"'"))
            return values
    return None


def tasks_at(root: Path) -> list[dict[str, Any]]:
    tasks = json_command(["taskmd", "list", "--format", "json"], root)
    snapshot = json_command(["taskmd", "snapshot", "--format", "json"], root)
    absolute_paths = {
        item["id"]: item.get("file_path", "") for item in snapshot.get("tasks", [])
    }
    for item in tasks:
        if absolute_paths.get(item["id"]):
            item["file_path"] = absolute_paths[item["id"]]
        resources = frontmatter_list(item.get("file_path", ""), "resources")
        item["resources"] = resources or []
        item["resources_declared"] = resources is not None
        item["merged_commit"] = git(
            root, "config", "--get", f"taskmd.task.{item['id']}.merged-commit", check=False
        )
        item["validated_commit"] = git(
            root,
            "config",
            "--get",
            f"taskmd.task.{item['id']}.validated-commit",
            check=False,
        )
    return tasks


def collect_state(root: Path, task_query: str, main_override: str | None) -> dict[str, Any]:
    root = Path(git(root, "rev-parse", "--show-toplevel"))
    main = main_override or default_branch(root)
    tasks = tasks_at(root)
    requested = json_command(
        ["taskmd", "get", task_query, "--exact", "--format", "json"], root
    )
    task_id = requested["id"]
    worktrees = worktree_records(root, main)
    task_views = list(tasks)
    for worktree in worktrees:
        worktree_root = Path(worktree["path"])
        if worktree_root == root:
            continue
        try:
            task_views.extend(tasks_at(worktree_root))
        except CollectionError:
            worktree["task_state_error"] = True
    merged_tasks = {item["id"]: item for item in tasks}
    status_rank = {
        "pending": 0,
        "blocked": 1,
        "in-progress": 2,
        "in-review": 3,
        "completed": 4,
        "cancelled": 4,
    }
    for item in task_views:
        current = merged_tasks.get(item["id"])
        if current is None or status_rank.get(item.get("status"), 0) > status_rank.get(
            current.get("status"), 0
        ):
            merged_tasks[item["id"]] = item
    tasks = list(merged_tasks.values())
    task = merged_tasks.get(task_id, requested)
    active_by_id: dict[str, dict[str, Any]] = {}
    for item in task_views:
        if item.get("status") == "in-progress" and item.get("id") != task_id:
            active_by_id[item["id"]] = item
    active = list(active_by_id.values())
    relevant_ids = {task_id, *(task.get("dependencies") or [])}
    relevant_ids.update(item["id"] for item in active)
    tasks_by_id = {item["id"]: item for item in tasks}
    scopes = {
        item_id: context_paths(root, tasks_by_id[item_id])
        for item_id in relevant_ids
        if item_id in tasks_by_id
    }
    tracks = json_command(["taskmd", "tracks", "--format", "json"], root)
    return {
        "project_root": str(root),
        "main": {
            "branch": main,
            "commit": git(root, "rev-parse", main),
        },
        "task": task,
        "tasks": tasks,
        "active_tasks": active,
        "worktrees": worktrees,
        "branches": branch_records(root, main),
        "scopes": scopes,
        "tracks": tracks,
    }


def path_overlaps(path: str, scope_path: str) -> bool:
    path = path.strip("/")
    scope_path = scope_path.strip("/")
    if not path or not scope_path:
        return True
    return (
        path == scope_path
        or path.startswith(scope_path + "/")
        or scope_path.startswith(path + "/")
    )


def scope_quality(task: dict[str, Any], scope: dict[str, Any]) -> list[str]:
    problems: list[str] = []
    touches = task.get("touches") or []
    paths = scope.get("paths") or []
    if not touches:
        problems.append("touches is missing")
    if not paths:
        problems.append("touches resolves to no paths")
    if scope.get("unresolved"):
        problems.append("touches contains unresolved non-path scopes")
    broad = {".", "./", "/", "*", "**", "./**"}
    if any(path.strip() in broad or "*" in path for path in paths):
        problems.append("touches contains a broad or wildcard path")
    return problems


def mapped_worktree(
    task_id: str, worktrees: list[dict[str, Any]]
) -> tuple[dict[str, Any] | None, bool]:
    explicit = [item for item in worktrees if item.get("task_id") == task_id]
    if len(explicit) == 1:
        return explicit[0], False
    if len(explicit) > 1:
        return None, True
    token = re.compile(rf"(^|[^A-Za-z0-9]){re.escape(task_id)}([^A-Za-z0-9]|$)")
    inferred = [
        item
        for item in worktrees
        if token.search(item.get("branch", "")) or token.search(item.get("path", ""))
    ]
    if len(inferred) == 1:
        return inferred[0], False
    return None, len(inferred) > 1


def task_on_main(state: dict[str, Any], task: dict[str, Any]) -> bool:
    override = state.get("merged_to_main", {}).get(task.get("id"))
    if override is not None:
        return bool(override)
    root = Path(state["project_root"])
    main = state["main"]["branch"]
    mapped = [
        branch
        for branch in state.get("branches", [])
        if branch.get("task_id") == task.get("id")
    ]
    if len(mapped) == 1:
        return bool(
            mapped[0].get("head_merged_to_main")
            and mapped[0].get("validated_commit") == mapped[0].get("head")
        )
    if len(mapped) > 1:
        return False
    merged_commit = task.get("merged_commit", "")
    validated_commit = task.get("validated_commit", "")
    return bool(
        merged_commit
        and merged_commit == validated_commit
        and commit_is_ancestor(root, merged_commit, main)
    )


def stack_base_problems(
    dependency: dict[str, Any] | None,
    worktree: dict[str, Any] | None,
    ambiguous: bool,
) -> list[str]:
    problems = []
    if dependency is None:
        return ["dependency is missing"]
    if dependency.get("status") not in {"in-progress", "completed"}:
        problems.append("dependency is not in a stackable task state")
    if ambiguous:
        problems.append("dependency maps to multiple worktrees")
    if worktree is None:
        problems.append("dependency unmerged branch is unavailable")
        return problems
    if not worktree.get("clean"):
        problems.append("dependency worktree is dirty")
    if worktree.get("commits_ahead", 0) < 1:
        problems.append("dependency work is not committed on its branch")
    if not worktree.get("frozen"):
        problems.append("dependency branch is not explicitly frozen")
    if worktree.get("validated_commit") != worktree.get("head"):
        problems.append("dependency HEAD is not recorded as validated")
    return problems


def track_for(task_id: str, tracks: dict[str, Any]) -> dict[str, Any] | None:
    for track in tracks.get("tracks", []):
        if any(item.get("id") == task_id for item in track.get("tasks", [])):
            return {"id": track.get("id"), "scopes": track.get("scopes", [])}
    if any(item.get("id") == task_id for item in tracks.get("flexible", [])):
        return {"id": None, "scopes": [], "flexible": True}
    return None


def base_result(state: dict[str, Any]) -> dict[str, Any]:
    task = state["task"]
    task_id = task["id"]
    return {
        "schema_version": 1,
        "decision": "UNSAFE",
        "task": {
            "id": task_id,
            "title": task.get("title", ""),
            "status": task.get("status", ""),
            "file_path": task.get("file_path", ""),
            "touches": task.get("touches", []),
            "resources": task.get("resources", []),
            "resources_declared": task.get("resources_declared", False),
        },
        "proposal": {
            "base_branch": state["main"]["branch"],
            "base_commit": state["main"]["commit"],
            "worktree_name": f"task-{task_id}",
            "merge_order": [task_id],
            "post_merge": [],
            "validation": ["run every verification required by the task before completion"],
        },
        "evidence": {
            "dependencies": [],
            "active_tasks": [],
            "worktrees": state.get("worktrees", []),
            "declared_overlaps": [],
            "resource_overlaps": [],
            "actual_overlaps": [],
            "stack_base_changes": [],
            "scope_problems": [],
            "track": track_for(task_id, state.get("tracks", {})),
        },
        "reason": "",
        "requires_confirmation": True,
        "automatic_allowed": False,
        "waiting": {
            "self_waking": False,
            "resume_via": ["active agent goal polling", "external scheduler"],
        },
    }


def assess_decision(state: dict[str, Any], allow_stacked: bool = False) -> dict[str, Any]:
    result = base_result(state)
    task = state["task"]
    task_id = task["id"]
    tasks = {item["id"]: item for item in state.get("tasks", [])}
    worktrees = state.get("worktrees", [])
    target_scope = state.get("scopes", {}).get(task_id, {})
    problems = scope_quality(task, target_scope)
    result["evidence"]["scope_problems"] = problems

    if not SAFE_TASK_ID.fullmatch(task_id):
        result["reason"] = "task ID contains characters unsafe for branch or Git config use"
        return result
    if task.get("status") != "pending":
        result["reason"] = "requested task must be pending before queue execution"
        return result
    if problems:
        result["reason"] = "scope evidence is missing, broad, stale, or ambiguous"
        return result
    if any(worktree.get("task_state_error") for worktree in worktrees):
        result["reason"] = "task state could not be inspected in every active worktree"
        return result

    target_paths = target_scope.get("paths", [])
    eligible_stack_paths: set[str] = set()
    for dependency_id in task.get("dependencies") or []:
        dependency = tasks.get(dependency_id)
        dependency_worktree, ambiguous = mapped_worktree(dependency_id, worktrees)
        if not stack_base_problems(dependency, dependency_worktree, ambiguous):
            eligible_stack_paths.add(dependency_worktree.get("path", ""))

    ambiguous_active = False
    for active in state.get("active_tasks", []):
        active_scope = state.get("scopes", {}).get(active["id"], {})
        active_problems = scope_quality(active, active_scope)
        active_worktree, ambiguous = mapped_worktree(active["id"], worktrees)
        if ambiguous or not active_worktree:
            active_problems.append("in-progress task does not map to exactly one worktree")
        stale_files = []
        if active_worktree:
            active_paths = active_scope.get("paths", [])
            stale_files = sorted(
                changed
                for changed in active_worktree.get("changed_files", [])
                if not any(path_overlaps(changed, path) for path in active_paths)
            )
            if stale_files:
                active_problems.append("actual changes fall outside declared touches")
        declared = sorted(set(task.get("touches", [])) & set(active.get("touches", [])))
        resolved = sorted(
            {
                f"{target} ↔ {active_path}"
                for target in target_paths
                for active_path in active_scope.get("paths", [])
                if path_overlaps(target, active_path)
            }
        )
        if declared or resolved:
            result["evidence"]["declared_overlaps"].append(
                {"task_id": active["id"], "scopes": declared, "paths": resolved}
            )
        resources = sorted(
            set(task.get("resources", [])) & set(active.get("resources", []))
        )
        if resources:
            result["evidence"]["resource_overlaps"].append(
                {"task_id": active["id"], "resources": resources}
            )
        result["evidence"]["active_tasks"].append(
            {
                "id": active["id"],
                "touches": active.get("touches", []),
                "scope_problems": active_problems,
                "worktree": active_worktree,
                "stale_files": stale_files,
            }
        )
        if active_problems:
            ambiguous_active = True

    for worktree in worktrees:
        overlaps = sorted(
            changed
            for changed in worktree.get("changed_files", [])
            if any(path_overlaps(changed, target) for target in target_paths)
        )
        if overlaps:
            evidence = {
                "branch": worktree.get("branch", ""),
                "path": worktree.get("path", ""),
                "files": overlaps,
            }
            if worktree.get("path", "") in eligible_stack_paths:
                result["evidence"]["stack_base_changes"].append(evidence)
            else:
                result["evidence"]["actual_overlaps"].append(evidence)

    if result["evidence"]["actual_overlaps"]:
        result["decision"] = "WAIT"
        result["reason"] = "actual mutable files overlap the requested task scope"
        return result
    if result["evidence"]["declared_overlaps"]:
        result["decision"] = "WAIT"
        result["reason"] = "declared scopes overlap an in-progress task"
        return result
    if result["evidence"]["resource_overlaps"]:
        result["decision"] = "WAIT"
        result["reason"] = "operational resources overlap an in-progress task"
        return result
    if ambiguous_active:
        result["reason"] = "an in-progress task has ambiguous scope evidence"
        return result

    dependencies = task.get("dependencies") or []
    unresolved: list[dict[str, Any]] = []
    for dependency_id in dependencies:
        dependency = tasks.get(dependency_id)
        evidence = {"id": dependency_id, "state": "missing"}
        if dependency:
            worktree, ambiguous = mapped_worktree(dependency_id, worktrees)
            merged = task_on_main(state, dependency)
            evidence.update(
                {
                    "status": dependency.get("status"),
                    "merged_to_main": merged,
                    "worktree": worktree,
                    "ambiguous_worktree": ambiguous,
                }
            )
            if not merged:
                unresolved.append(evidence)
        else:
            unresolved.append(evidence)
        result["evidence"]["dependencies"].append(evidence)

    if not unresolved:
        result["decision"] = "START_FROM_MAIN"
        result["reason"] = "dependencies are present on main and no mutable scope overlaps exist"
        result["requires_confirmation"] = True
        resource_evidence_complete = task.get("resources_declared", False) and all(
            active.get("resources_declared", False)
            for active in state.get("active_tasks", [])
        )
        result["automatic_allowed"] = resource_evidence_complete
        if not resource_evidence_complete:
            result["proposal"]["validation"].append(
                "automatic execution is disabled because operational resources were not explicitly declared"
            )
        return result

    if len(unresolved) != 1:
        result["decision"] = "WAIT"
        result["reason"] = "multiple or missing dependencies are not present on main"
        return result

    dependency = unresolved[0]
    worktree = dependency.get("worktree")
    stack_problems = stack_base_problems(
        tasks.get(dependency["id"]), worktree, dependency.get("ambiguous_worktree", False)
    )
    if stack_problems:
        result["decision"] = "WAIT"
        result["reason"] = stack_problems[0]
        return result
    if not allow_stacked:
        result["decision"] = "WAIT"
        result["reason"] = "stacking is eligible but requires explicit user authorization"
        return result

    result["decision"] = "START_STACKED"
    result["reason"] = "one clean, committed, validated, frozen dependency is eligible as a base"
    result["proposal"].update(
        {
            "base_branch": worktree["branch"],
            "base_commit": worktree["head"],
            "merge_order": [dependency["id"], task_id],
            "post_merge": [
                f"rebase task-{task_id} onto updated {state['main']['branch']}",
                "rerun all validation required by the stacked task",
            ],
            "validation": [
                "base validation is pinned to the recorded base commit",
                "rerun every stacked-task verification after rebasing onto updated main",
            ],
        }
    )
    return result


def assess(state: dict[str, Any], allow_stacked: bool = False) -> dict[str, Any]:
    result = assess_decision(state, allow_stacked)
    approval_payload = {
        "task_id": result.get("task", {}).get("id"),
        "decision": result.get("decision"),
        "proposal": result.get("proposal"),
    }
    encoded = json.dumps(
        approval_payload, sort_keys=True, separators=(",", ":")
    ).encode()
    result["approval_token"] = hashlib.sha256(encoded).hexdigest()
    return result


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser()
    parser.add_argument("task", nargs="?", help="task ID, exact title, or file path")
    parser.add_argument("--project-root", default=".")
    parser.add_argument("--main")
    parser.add_argument("--allow-stacked", action="store_true")
    parser.add_argument("--state-file", help="use pre-collected JSON evidence")
    parser.add_argument("--pretty", action="store_true")
    return parser.parse_args()


def main() -> int:
    args = parse_args()
    try:
        if args.state_file:
            state = json.loads(Path(args.state_file).read_text())
        elif args.task:
            state = collect_state(Path(args.project_root), args.task, args.main)
        else:
            raise CollectionError("task is required unless --state-file is used")
        result = assess(state, allow_stacked=args.allow_stacked)
    except (CollectionError, OSError, KeyError, json.JSONDecodeError) as error:
        result = {
            "schema_version": 1,
            "decision": "UNSAFE",
            "reason": str(error),
            "requires_confirmation": True,
            "automatic_allowed": False,
        }
    if result["decision"] not in DECISIONS:
        raise AssertionError("assessment emitted an invalid decision")
    json.dump(result, sys.stdout, indent=2 if args.pretty else None, sort_keys=True)
    sys.stdout.write("\n")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
