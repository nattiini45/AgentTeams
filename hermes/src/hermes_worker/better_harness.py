"""Weekly better-harness self-review trigger for the managed Hermes worker.

The worker daemon owns the native periodic trigger for the weekly harness
self-review. A supervised loop wakes on a short interval and invokes the
builtin skill's ``run-weekly.sh``; that script enforces the deterministic
7-day cadence guard, so the loop only actually collects evidence once a week
no matter how often it wakes. State (``$HOME/.better-harness/state.json``) is
persisted to object storage by the worker's normal push loop, so the cadence
survives container restarts.

The loop only *collects evidence* (read-only). The agent authors findings from
that evidence on its next wake (per the workspace AGENTS.md pointer) and
reports them to the Manager.
"""

from __future__ import annotations

import asyncio
import logging
import os
from pathlib import Path

logger = logging.getLogger(__name__)

# How often the loop wakes to check the cadence guard. The 7-day guard inside
# run-weekly.sh makes the effective cadence weekly; this interval only controls
# how promptly a due review is noticed after a restart.
DEFAULT_CHECK_INTERVAL_SECONDS = 3600


def _better_harness_enabled() -> bool:
    value = os.getenv("AGENTTEAMS_BETTER_HARNESS_ENABLED", "").strip().lower()
    return value not in {"0", "false", "no"}


def _run_weekly_script(hermes_home: Path) -> Path:
    return hermes_home / "skills" / "better-harness" / "scripts" / "run-weekly.sh"


async def run_better_harness_weekly_loop(
    hermes_home: Path,
    *,
    worker_name: str,
    check_interval: float = DEFAULT_CHECK_INTERVAL_SECONDS,
) -> None:
    """Periodically invoke the builtin better-harness weekly-review script.

    The script is idempotent and self-guarding (7-day cadence); this loop just
    ensures a due review is triggered promptly after the worker starts and at
    a steady cadence thereafter. Best-effort: any failure is logged and the
    loop continues.
    """

    if not _better_harness_enabled():
        logger.info(
            "better-harness weekly loop disabled component=better-harness worker=%s",
            worker_name,
        )
        return

    script = _run_weekly_script(hermes_home)
    logger.info(
        "better-harness weekly loop started component=better-harness worker=%s interval_seconds=%s script=%s",
        worker_name,
        check_interval,
        script,
    )

    try:
        while True:
            if not script.is_file():
                logger.debug(
                    "better-harness script not yet present component=better-harness worker=%s script=%s",
                    worker_name,
                    script,
                )
            else:
                try:
                    proc = await asyncio.create_subprocess_exec(
                        "/bin/bash",
                        str(script),
                        stdout=asyncio.subprocess.DEVNULL,
                        stderr=asyncio.subprocess.PIPE,
                    )
                    _, stderr = await proc.communicate()
                    if proc.returncode != 0:
                        logger.warning(
                            "better-harness weekly run exited non-zero component=better-harness worker=%s returncode=%s stderr=%r",
                            worker_name,
                            proc.returncode,
                            (stderr or b"").decode("utf-8", "replace")[:300],
                        )
                except Exception as exc:
                    logger.warning(
                        "better-harness weekly run failed component=better-harness worker=%s error_type=%s",
                        worker_name,
                        type(exc).__name__,
                    )
            await asyncio.sleep(check_interval)
    except asyncio.CancelledError:
        logger.info(
            "better-harness weekly loop stopped component=better-harness worker=%s",
            worker_name,
        )
        raise
