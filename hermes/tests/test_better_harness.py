"""Unit tests for the weekly better-harness trigger loop (hermes worker)."""

from __future__ import annotations

import asyncio
from pathlib import Path

import pytest

from hermes_worker.better_harness import (
    _better_harness_enabled,
    _run_weekly_script,
    run_better_harness_weekly_loop,
)


def test_better_harness_enabled_default_true(monkeypatch: pytest.MonkeyPatch) -> None:
    monkeypatch.delenv("AGENTTEAMS_BETTER_HARNESS_ENABLED", raising=False)
    assert _better_harness_enabled() is True


@pytest.mark.parametrize("value", ["0", "false", "False", "FALSE", "no", "NO"])
def test_better_harness_enabled_falsy(monkeypatch: pytest.MonkeyPatch, value: str) -> None:
    monkeypatch.setenv("AGENTTEAMS_BETTER_HARNESS_ENABLED", value)
    assert _better_harness_enabled() is False


def test_run_weekly_script_path() -> None:
    hermes_home = Path("/tmp/worker-a/.hermes")
    assert _run_weekly_script(hermes_home) == hermes_home / "skills" / "better-harness" / "scripts" / "run-weekly.sh"


def test_loop_returns_immediately_when_disabled(tmp_path: Path, monkeypatch: pytest.MonkeyPatch) -> None:
    monkeypatch.setenv("AGENTTEAMS_BETTER_HARNESS_ENABLED", "0")
    asyncio.run(asyncio.wait_for(run_better_harness_weekly_loop(tmp_path, worker_name="worker-a"), timeout=5))


def test_loop_skips_when_script_missing(tmp_path: Path, monkeypatch: pytest.MonkeyPatch) -> None:
    monkeypatch.delenv("AGENTTEAMS_BETTER_HARNESS_ENABLED", raising=False)

    async def _run_once_then_cancel() -> None:
        task = asyncio.create_task(
            run_better_harness_weekly_loop(tmp_path, worker_name="worker-a", check_interval=3600)
        )
        await asyncio.sleep(0.1)
        task.cancel()
        try:
            await task
        except asyncio.CancelledError:
            pass

    asyncio.run(_run_once_then_cancel())


def test_loop_invokes_script_when_present(tmp_path: Path, monkeypatch: pytest.MonkeyPatch) -> None:
    monkeypatch.delenv("AGENTTEAMS_BETTER_HARNESS_ENABLED", raising=False)
    script = _run_weekly_script(tmp_path)
    script.parent.mkdir(parents=True, exist_ok=True)
    script.write_text("#!/bin/bash\nexit 0\n", encoding="utf-8")

    calls: list[list[str]] = []

    class _FakeProc:
        returncode = 0

        async def communicate(self):
            return (b"", b"")

    async def fake_exec(*args, **kwargs):
        calls.append([str(a) for a in args])
        return _FakeProc()

    monkeypatch.setattr(asyncio, "create_subprocess_exec", fake_exec)

    async def _run_once_then_cancel() -> None:
        task = asyncio.create_task(
            run_better_harness_weekly_loop(tmp_path, worker_name="worker-a", check_interval=3600)
        )
        await asyncio.sleep(0.2)
        task.cancel()
        try:
            await task
        except asyncio.CancelledError:
            pass

    asyncio.run(_run_once_then_cancel())

    assert calls, "expected the loop to invoke run-weekly.sh"
    assert calls[0][0] == "/bin/bash"
    assert calls[0][1] == str(script)
