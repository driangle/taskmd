from __future__ import annotations

import json
import os
import subprocess
import sys
import tempfile
import unittest
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))

import safe_queue
import execute_safe_queue


ROOT = Path(__file__).resolve().parent
EXECUTE = ROOT / "execute_safe_queue.py"


def base_state() -> dict:
    task = {
        "id": "200",
        "title": "Independent",
        "status": "pending",
        "touches": ["app"],
        "dependencies": [],
        "file_path": "tasks/200-independent.md",
        "resources": [],
        "resources_declared": True,
    }
    return {
        "project_root": "/repo",
        "main": {"branch": "main", "commit": "main-sha"},
        "task": task,
        "tasks": [task],
        "active_tasks": [],
        "worktrees": [],
        "branches": [],
        "scopes": {"200": {"paths": ["src/app"], "missing": []}},
        "tracks": {"tracks": [{"id": 1, "tasks": [{"id": "200"}], "scopes": ["app"]}]},
    }


class AssessmentTests(unittest.TestCase):
    def test_independent_task_starts_from_main(self) -> None:
        result = safe_queue.assess(base_state())
        self.assertEqual("START_FROM_MAIN", result["decision"])
        self.assertTrue(result["automatic_allowed"])

    def test_declared_overlap_waits(self) -> None:
        state = base_state()
        active = {
            "id": "100",
            "status": "in-progress",
            "touches": ["app"],
            "dependencies": [],
        }
        state["active_tasks"] = [active]
        state["tasks"].append(active)
        state["scopes"]["100"] = {"paths": ["src/app"], "missing": []}
        state["worktrees"] = [
            {
                "branch": "task-100",
                "path": "/repo.task-100",
                "head": "active-sha",
                "clean": True,
                "changed_files": [],
                "task_id": "100",
            }
        ]
        result = safe_queue.assess(state)
        self.assertEqual("WAIT", result["decision"])
        self.assertTrue(result["evidence"]["declared_overlaps"])

    def test_actual_file_overlap_waits_even_without_declared_overlap(self) -> None:
        state = base_state()
        state["worktrees"] = [
            {
                "branch": "other",
                "path": "/repo.other",
                "head": "other-sha",
                "clean": False,
                "changed_files": ["src/app/handler.go"],
            }
        ]
        result = safe_queue.assess(state)
        self.assertEqual("WAIT", result["decision"])
        self.assertTrue(result["evidence"]["actual_overlaps"])

    def test_operational_resource_overlap_waits(self) -> None:
        state = base_state()
        state["task"]["resources"] = ["port:3000"]
        active = {
            "id": "100",
            "status": "in-progress",
            "touches": ["base"],
            "resources": ["port:3000"],
            "resources_declared": True,
            "dependencies": [],
        }
        state["active_tasks"] = [active]
        state["tasks"].append(active)
        state["scopes"]["100"] = {"paths": ["src/base"], "missing": []}
        state["worktrees"] = [
            {
                "branch": "task-100",
                "path": "/repo.task-100",
                "head": "active-sha",
                "clean": True,
                "changed_files": [],
                "task_id": "100",
            }
        ]
        result = safe_queue.assess(state)
        self.assertEqual("WAIT", result["decision"])
        self.assertTrue(result["evidence"]["resource_overlaps"])

    def test_automatic_mode_requires_explicit_resource_declaration(self) -> None:
        state = base_state()
        state["task"]["resources_declared"] = False
        result = safe_queue.assess(state)
        self.assertEqual("START_FROM_MAIN", result["decision"])
        self.assertFalse(result["automatic_allowed"])

    def test_missing_or_broad_scope_is_unsafe(self) -> None:
        for touches, paths in (([], []), (["all"], ["."]), (["wild"], ["src/**"])):
            with self.subTest(touches=touches):
                state = base_state()
                state["task"]["touches"] = touches
                state["scopes"]["200"]["paths"] = paths
                self.assertEqual("UNSAFE", safe_queue.assess(state)["decision"])

    def test_unsafe_task_id_is_rejected(self) -> None:
        state = base_state()
        state["task"]["id"] = "200\nbad"
        self.assertEqual("UNSAFE", safe_queue.assess(state)["decision"])

    def test_unreadable_worktree_task_state_is_unsafe(self) -> None:
        state = base_state()
        state["worktrees"] = [
            {
                "branch": "other",
                "path": "/repo.other",
                "head": "other-sha",
                "clean": True,
                "changed_files": [],
                "task_state_error": True,
            }
        ]
        result = safe_queue.assess(state)
        self.assertEqual("UNSAFE", result["decision"])

    def test_dirty_dependency_waits(self) -> None:
        state = self.stacked_state()
        state["worktrees"][0]["clean"] = False
        self.assertEqual(
            "WAIT", safe_queue.assess(state, allow_stacked=True)["decision"]
        )

    def test_frozen_validated_dependency_can_start_stacked(self) -> None:
        state = self.stacked_state()
        without_approval = safe_queue.assess(state)
        self.assertEqual("WAIT", without_approval["decision"])
        result = safe_queue.assess(state, allow_stacked=True)
        self.assertEqual("START_STACKED", result["decision"])
        self.assertEqual(["100", "200"], result["proposal"]["merge_order"])
        self.assertEqual("dep-sha", result["proposal"]["base_commit"])
        self.assertTrue(result["proposal"]["post_merge"])

    def test_completed_but_unmerged_dependency_waits_without_branch(self) -> None:
        state = self.stacked_state()
        state["worktrees"] = []
        self.assertEqual(
            "WAIT", safe_queue.assess(state, allow_stacked=True)["decision"]
        )

    def test_merged_dependency_starts_from_main(self) -> None:
        state = self.stacked_state()
        state["merged_to_main"] = {"100": True}
        result = safe_queue.assess(state)
        self.assertEqual("START_FROM_MAIN", result["decision"])

    def test_completed_dependency_without_merge_evidence_waits(self) -> None:
        state = self.stacked_state()
        state["worktrees"] = []
        state.pop("merged_to_main")
        result = safe_queue.assess(state)
        self.assertEqual("WAIT", result["decision"])

    def test_mapped_dependency_branch_proves_merge_to_main(self) -> None:
        state = self.stacked_state()
        state["worktrees"] = []
        state.pop("merged_to_main")
        state["branches"] = [
            {
                "branch": "task-100",
                "head": "dep-sha",
                "task_id": "100",
                "head_merged_to_main": True,
                "validated_commit": "dep-sha",
            }
        ]
        result = safe_queue.assess(state)
        self.assertEqual("START_FROM_MAIN", result["decision"])

    def test_explicit_merged_commit_supports_ignored_task_files(self) -> None:
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            subprocess.run(["git", "init", "-b", "main"], cwd=root, check=True)
            subprocess.run(
                ["git", "config", "user.email", "test@example.com"], cwd=root, check=True
            )
            subprocess.run(["git", "config", "user.name", "Test"], cwd=root, check=True)
            (root / "README.md").write_text("fixture\n")
            subprocess.run(["git", "add", "README.md"], cwd=root, check=True)
            subprocess.run(["git", "commit", "-m", "fixture"], cwd=root, check=True)
            head = subprocess.run(
                ["git", "rev-parse", "HEAD"],
                cwd=root,
                text=True,
                capture_output=True,
                check=True,
            ).stdout.strip()
            state = {
                "project_root": str(root),
                "main": {"branch": "main", "commit": head},
                "branches": [],
            }
            dependency = {
                "id": "100",
                "merged_commit": head,
                "validated_commit": head,
            }
            self.assertTrue(safe_queue.task_on_main(state, dependency))

    def test_active_task_with_stale_touches_is_unsafe(self) -> None:
        state = base_state()
        active = {
            "id": "100",
            "status": "in-progress",
            "touches": ["base"],
            "dependencies": [],
        }
        state["active_tasks"] = [active]
        state["tasks"].append(active)
        state["scopes"]["100"] = {"paths": ["src/base"], "missing": []}
        state["worktrees"] = [
            {
                "branch": "task-100",
                "path": "/repo.task-100",
                "head": "active-sha",
                "clean": False,
                "changed_files": ["src/undeclared/file.go"],
                "task_id": "100",
            }
        ]
        result = safe_queue.assess(state)
        self.assertEqual("UNSAFE", result["decision"])
        self.assertIn(
            "actual changes fall outside declared touches",
            result["evidence"]["active_tasks"][0]["scope_problems"],
        )

    @staticmethod
    def stacked_state() -> dict:
        state = base_state()
        dependency = {
            "id": "100",
            "title": "Base",
            "status": "completed",
            "touches": ["base"],
            "dependencies": [],
            "file_path": "tasks/100-base.md",
        }
        state["task"]["dependencies"] = ["100"]
        state["tasks"].append(dependency)
        state["scopes"]["100"] = {"paths": ["src/base"], "missing": []}
        state["worktrees"] = [
            {
                "branch": "task-100",
                "path": "/repo.task-100",
                "head": "dep-sha",
                "clean": True,
                "commits_ahead": 1,
                "changed_files": ["src/app/base.go"],
                "task_id": "100",
                "frozen": True,
                "validated_commit": "dep-sha",
            }
        ]
        state["merged_to_main"] = {"100": False}
        return state


class ExecutionTests(unittest.TestCase):
    def test_failed_worktree_creation_does_not_claim_task(self) -> None:
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            state = base_state()
            state["project_root"] = str(root)
            state_file = root / "state.json"
            state_file.write_text(json.dumps(state))
            approval_token = safe_queue.assess(state)["approval_token"]
            bin_dir = root / "bin"
            bin_dir.mkdir()
            log = root / "commands.log"
            self.write_executable(
                bin_dir / "wt",
                f"#!/bin/sh\necho wt >> {log}\necho failed >&2\nexit 42\n",
            )
            self.write_executable(
                bin_dir / "taskmd",
                f"#!/bin/sh\necho taskmd >> {log}\nexit 0\n",
            )
            result = subprocess.run(
                [
                    str(EXECUTE),
                    "--state-file",
                    str(state_file),
                    "--project-root",
                    str(root),
                    "--approved",
                    "--expected-decision",
                    "START_FROM_MAIN",
                    "--expected-base-branch",
                    "main",
                    "--expected-base-commit",
                    "main-sha",
                    "--expected-approval-token",
                    approval_token,
                ],
                text=True,
                capture_output=True,
                env={
                    **os.environ,
                    "PATH": f"{bin_dir}:{os.environ['PATH']}",
                    "TASKMD_SAFE_QUEUE_TESTING": "1",
                },
                check=False,
            )
            self.assertNotEqual(0, result.returncode)
            self.assertEqual(["wt"], log.read_text().splitlines())

    def test_success_copies_only_untracked_task_before_claim(self) -> None:
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory) / "repo"
            root.mkdir()
            subprocess.run(["git", "init", "-b", "main"], cwd=root, check=True)
            subprocess.run(
                ["git", "config", "user.email", "test@example.com"], cwd=root, check=True
            )
            subprocess.run(["git", "config", "user.name", "Test"], cwd=root, check=True)
            (root / "README.md").write_text("fixture\n")
            subprocess.run(["git", "add", "README.md"], cwd=root, check=True)
            subprocess.run(["git", "commit", "-m", "fixture"], cwd=root, check=True)

            task_file = root / "tasks" / "200-independent.md"
            task_file.parent.mkdir()
            task_file.write_text("---\nid: \"200\"\nstatus: pending\n---\n")
            runtime_file = root / "runtime.state"
            runtime_file.write_text("do not copy\n")

            state = base_state()
            state["project_root"] = str(root)
            state["main"]["commit"] = subprocess.run(
                ["git", "rev-parse", "HEAD"],
                cwd=root,
                text=True,
                capture_output=True,
                check=True,
            ).stdout.strip()
            state["task"]["file_path"] = str(task_file)
            state_file = root / "state.json"
            state_file.write_text(json.dumps(state))
            approval_token = safe_queue.assess(state)["approval_token"]

            bin_dir = root / "bin"
            bin_dir.mkdir()
            worktree = Path(directory) / "repo.task-200"
            log = root / "commands.log"
            self.write_executable(
                bin_dir / "wt",
                "#!/bin/sh\n"
                f"git -C {root} worktree add -b task-200 {worktree} main >/dev/null\n",
            )
            self.write_executable(
                bin_dir / "taskmd",
                "#!/bin/sh\n"
                f"echo \"$PWD $*\" >> {log}\n",
            )

            result = subprocess.run(
                [
                    str(EXECUTE),
                    "--state-file",
                    str(state_file),
                    "--project-root",
                    str(root),
                    "--approved",
                    "--expected-decision",
                    "START_FROM_MAIN",
                    "--expected-base-branch",
                    "main",
                    "--expected-base-commit",
                    state["main"]["commit"],
                    "--expected-approval-token",
                    approval_token,
                    "--owner",
                    "agent",
                ],
                text=True,
                capture_output=True,
                env={
                    **os.environ,
                    "PATH": f"{bin_dir}:{os.environ['PATH']}",
                    "TASKMD_SAFE_QUEUE_TESTING": "1",
                },
                check=False,
            )
            self.assertEqual(0, result.returncode, result.stdout + result.stderr)
            self.assertIn(
                f"{root} set 200 --status in-progress --owner agent", log.read_text()
            )
            self.assertTrue((worktree / "tasks" / task_file.name).exists())
            self.assertFalse((worktree / runtime_file.name).exists())
            self.assertEqual(
                "main",
                subprocess.run(
                    ["git", "config", "--get", "branch.task-200.taskmd-base-branch"],
                    cwd=root,
                    text=True,
                    capture_output=True,
                    check=True,
                ).stdout.strip(),
            )
            self.assertEqual(
                state["main"]["commit"],
                subprocess.run(
                    ["git", "config", "--get", "branch.task-200.taskmd-base-commit"],
                    cwd=root,
                    text=True,
                    capture_output=True,
                    check=True,
                ).stdout.strip(),
            )

    def test_changed_proposal_is_refused_before_worktree_creation(self) -> None:
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            state = base_state()
            state["project_root"] = str(root)
            state_file = root / "state.json"
            state_file.write_text(json.dumps(state))
            approval_token = safe_queue.assess(state)["approval_token"]
            bin_dir = root / "bin"
            bin_dir.mkdir()
            log = root / "commands.log"
            self.write_executable(
                bin_dir / "wt",
                f"#!/bin/sh\necho wt >> {log}\nexit 0\n",
            )
            result = subprocess.run(
                [
                    str(EXECUTE),
                    "--state-file",
                    str(state_file),
                    "--project-root",
                    str(root),
                    "--approved",
                    "--expected-decision",
                    "START_FROM_MAIN",
                    "--expected-base-branch",
                    "main",
                    "--expected-base-commit",
                    "previous-main-sha",
                    "--expected-approval-token",
                    approval_token,
                ],
                text=True,
                capture_output=True,
                env={
                    **os.environ,
                    "PATH": f"{bin_dir}:{os.environ['PATH']}",
                    "TASKMD_SAFE_QUEUE_TESTING": "1",
                },
                check=False,
            )
            self.assertNotEqual(0, result.returncode)
            self.assertFalse(log.exists())
            self.assertIn("no longer matches", result.stdout)

    def test_full_proposal_token_mismatch_is_refused(self) -> None:
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            state = base_state()
            state["project_root"] = str(root)
            state_file = root / "state.json"
            state_file.write_text(json.dumps(state))
            bin_dir = root / "bin"
            bin_dir.mkdir()
            log = root / "commands.log"
            self.write_executable(
                bin_dir / "wt",
                f"#!/bin/sh\necho wt >> {log}\nexit 0\n",
            )
            result = subprocess.run(
                [
                    str(EXECUTE),
                    "--state-file",
                    str(state_file),
                    "--project-root",
                    str(root),
                    "--approved",
                    "--expected-decision",
                    "START_FROM_MAIN",
                    "--expected-base-branch",
                    "main",
                    "--expected-base-commit",
                    "main-sha",
                    "--expected-approval-token",
                    "0" * 64,
                ],
                text=True,
                capture_output=True,
                env={
                    **os.environ,
                    "PATH": f"{bin_dir}:{os.environ['PATH']}",
                    "TASKMD_SAFE_QUEUE_TESTING": "1",
                },
                check=False,
            )
            self.assertNotEqual(0, result.returncode)
            self.assertFalse(log.exists())
            self.assertIn("full explicitly approved proposal", result.stdout)

    def test_concurrent_execution_lock_refuses_before_assessment(self) -> None:
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            (root / "taskmd-safe-queue.lock").mkdir()
            result = subprocess.run(
                [
                    str(EXECUTE),
                    "--state-file",
                    str(root / "unused.json"),
                    "--project-root",
                    str(root),
                    "--approved",
                ],
                text=True,
                capture_output=True,
                env={**os.environ, "TASKMD_SAFE_QUEUE_TESTING": "1"},
                check=False,
            )
            self.assertNotEqual(0, result.returncode)
            self.assertIn("already in progress", result.stdout)

    def test_task_path_outside_project_is_safely_rejected(self) -> None:
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory) / "repo"
            root.mkdir()
            outside = Path(directory) / "outside-task.md"
            outside.write_text("task\n")
            with self.assertRaisesRegex(
                safe_queue.CollectionError, "outside the project root"
            ):
                execute_safe_queue.untracked_task_path(root, str(outside))

    @staticmethod
    def write_executable(path: Path, content: str) -> None:
        path.write_text(content)
        path.chmod(0o755)


if __name__ == "__main__":
    unittest.main()
