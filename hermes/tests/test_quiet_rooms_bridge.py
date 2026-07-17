"""Tests for AGENTTEAMS_VERBOSE_ROOMS -> MATRIX_FILTER_TOOL_MESSAGES/THINKING env
derivation in hermes_worker.bridge._matrix_env.

Phase 5b quiet-by-default: unset env derives filters ``"true"`` (suppress).
AGENTTEAMS_VERBOSE_ROOMS truthy derives ``"false"`` (verbose). AGENTTEAMS_QUIET_ROOMS
is deprecated with inverted legacy semantics.
"""
from __future__ import annotations

import logging

from hermes_worker.bridge import _matrix_env


def _make_cfg():
    return {
        "channels": {
            "matrix": {
                "enabled": True,
                "homeserver": "http://hiclaw-controller:6167",
                "accessToken": "tok",
                "userId": "@alice:matrix.local",
            }
        },
    }


def test_verbose_rooms_unset_defaults_filters_true(monkeypatch):
    monkeypatch.delenv("AGENTTEAMS_VERBOSE_ROOMS", raising=False)
    monkeypatch.delenv("AGENTTEAMS_QUIET_ROOMS", raising=False)

    env = _matrix_env(_make_cfg())

    assert env["MATRIX_FILTER_TOOL_MESSAGES"] == "true"
    assert env["MATRIX_FILTER_THINKING"] == "true"


def test_verbose_rooms_truthy_derives_filters_false(monkeypatch):
    monkeypatch.setenv("AGENTTEAMS_VERBOSE_ROOMS", "true")
    monkeypatch.delenv("AGENTTEAMS_QUIET_ROOMS", raising=False)

    env = _matrix_env(_make_cfg())

    assert env["MATRIX_FILTER_TOOL_MESSAGES"] == "false"
    assert env["MATRIX_FILTER_THINKING"] == "false"


def test_verbose_rooms_accepts_common_truthy_spellings(monkeypatch):
    for value in ("1", "yes", "on", "TRUE"):
        monkeypatch.setenv("AGENTTEAMS_VERBOSE_ROOMS", value)
        monkeypatch.delenv("AGENTTEAMS_QUIET_ROOMS", raising=False)
        env = _matrix_env(_make_cfg())
        assert env["MATRIX_FILTER_TOOL_MESSAGES"] == "false", f"failed for {value!r}"
        assert env["MATRIX_FILTER_THINKING"] == "false", f"failed for {value!r}"


def test_quiet_rooms_deprecated_true_keeps_suppress(monkeypatch, caplog):
    monkeypatch.delenv("AGENTTEAMS_VERBOSE_ROOMS", raising=False)
    monkeypatch.setenv("AGENTTEAMS_QUIET_ROOMS", "true")

    with caplog.at_level(logging.WARNING):
        env = _matrix_env(_make_cfg())

    assert env["MATRIX_FILTER_TOOL_MESSAGES"] == "true"
    assert env["MATRIX_FILTER_THINKING"] == "true"
    assert any("AGENTTEAMS_QUIET_ROOMS is deprecated" in r.message for r in caplog.records)


def test_quiet_rooms_deprecated_false_enables_verbose(monkeypatch, caplog):
    monkeypatch.delenv("AGENTTEAMS_VERBOSE_ROOMS", raising=False)
    monkeypatch.setenv("AGENTTEAMS_QUIET_ROOMS", "false")

    with caplog.at_level(logging.WARNING):
        env = _matrix_env(_make_cfg())

    assert env["MATRIX_FILTER_TOOL_MESSAGES"] == "false"
    assert env["MATRIX_FILTER_THINKING"] == "false"
    assert any("AGENTTEAMS_QUIET_ROOMS is deprecated" in r.message for r in caplog.records)


def test_verbose_rooms_wins_when_both_set(monkeypatch):
    monkeypatch.setenv("AGENTTEAMS_VERBOSE_ROOMS", "true")
    monkeypatch.setenv("AGENTTEAMS_QUIET_ROOMS", "true")

    env = _matrix_env(_make_cfg())

    assert env["MATRIX_FILTER_TOOL_MESSAGES"] == "false"
    assert env["MATRIX_FILTER_THINKING"] == "false"
