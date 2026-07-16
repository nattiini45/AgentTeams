"""Tests for AGENTTEAMS_VERBOSE_ROOMS -> config.json root show_tool_details.

Phase 5b "quiet rooms" (default quiet): unset env leaves show_tool_details absent
so pydantic default False applies. AGENTTEAMS_VERBOSE_ROOMS truthy writes root
``show_tool_details: true``. AGENTTEAMS_QUIET_ROOMS is deprecated (inverted
legacy semantics) with a warning when read.
"""

import json
import logging
import tempfile
from pathlib import Path

from copaw_worker.bridge import bridge_openclaw_to_copaw


def _make_openclaw_cfg():
    return {
        "channels": {
            "matrix": {
                "enabled": True,
                "homeserver": "http://localhost:6167",
                "accessToken": "tok",
            }
        },
        "models": {
            "providers": {
                "gw": {
                    "baseUrl": "http://aigw:8080/v1",
                    "apiKey": "key123",
                    "models": [{"id": "qwen3.5-plus", "name": "qwen3.5-plus"}],
                }
            }
        },
        "agents": {"defaults": {"model": {"primary": "gw/qwen3.5-plus"}}},
    }


def _config_json_path(working_dir: Path) -> dict:
    with open(working_dir / "config.json") as f:
        return json.load(f)


def test_verbose_rooms_unset_leaves_show_tool_details_absent(monkeypatch):
    monkeypatch.delenv("AGENTTEAMS_VERBOSE_ROOMS", raising=False)
    monkeypatch.delenv("AGENTTEAMS_QUIET_ROOMS", raising=False)

    with tempfile.TemporaryDirectory() as tmpdir:
        working_dir = Path(tmpdir) / "agent"
        bridge_openclaw_to_copaw(_make_openclaw_cfg(), working_dir, profile="manager")
        config = _config_json_path(working_dir)

    assert "show_tool_details" not in config
    assert config["channels"]["matrix"]["filter_tool_messages"] is True
    assert config["channels"]["matrix"]["filter_thinking"] is True


def test_verbose_rooms_truthy_sets_root_show_tool_details_true(monkeypatch):
    monkeypatch.setenv("AGENTTEAMS_VERBOSE_ROOMS", "true")
    monkeypatch.delenv("AGENTTEAMS_QUIET_ROOMS", raising=False)

    with tempfile.TemporaryDirectory() as tmpdir:
        working_dir = Path(tmpdir) / "agent"
        bridge_openclaw_to_copaw(_make_openclaw_cfg(), working_dir, profile="manager")
        config = _config_json_path(working_dir)

    assert config["show_tool_details"] is True


def test_verbose_rooms_accepts_common_truthy_spellings(monkeypatch):
    for value in ("1", "yes", "on", "TRUE", "True"):
        monkeypatch.setenv("AGENTTEAMS_VERBOSE_ROOMS", value)
        monkeypatch.delenv("AGENTTEAMS_QUIET_ROOMS", raising=False)

        with tempfile.TemporaryDirectory() as tmpdir:
            working_dir = Path(tmpdir) / "agent"
            bridge_openclaw_to_copaw(
                _make_openclaw_cfg(), working_dir, profile="manager"
            )
            config = _config_json_path(working_dir)

        assert config["show_tool_details"] is True, f"failed for {value!r}"


def test_verbose_rooms_toggling_off_after_on_removes_root_flag(monkeypatch):
    with tempfile.TemporaryDirectory() as tmpdir:
        working_dir = Path(tmpdir) / "agent"

        monkeypatch.setenv("AGENTTEAMS_VERBOSE_ROOMS", "true")
        bridge_openclaw_to_copaw(_make_openclaw_cfg(), working_dir, profile="manager")
        enabled = _config_json_path(working_dir)
        assert enabled["show_tool_details"] is True

        monkeypatch.delenv("AGENTTEAMS_VERBOSE_ROOMS", raising=False)
        bridge_openclaw_to_copaw(_make_openclaw_cfg(), working_dir, profile="manager")
        disabled = _config_json_path(working_dir)

    assert "show_tool_details" not in disabled


def test_quiet_rooms_deprecated_true_keeps_quiet(monkeypatch, caplog):
    monkeypatch.delenv("AGENTTEAMS_VERBOSE_ROOMS", raising=False)
    monkeypatch.setenv("AGENTTEAMS_QUIET_ROOMS", "true")

    with caplog.at_level(logging.WARNING):
        with tempfile.TemporaryDirectory() as tmpdir:
            working_dir = Path(tmpdir) / "agent"
            bridge_openclaw_to_copaw(
                _make_openclaw_cfg(), working_dir, profile="manager"
            )
            config = _config_json_path(working_dir)

    assert "show_tool_details" not in config
    assert any("AGENTTEAMS_QUIET_ROOMS is deprecated" in r.message for r in caplog.records)


def test_quiet_rooms_deprecated_false_enables_verbose(monkeypatch, caplog):
    monkeypatch.delenv("AGENTTEAMS_VERBOSE_ROOMS", raising=False)
    monkeypatch.setenv("AGENTTEAMS_QUIET_ROOMS", "false")

    with caplog.at_level(logging.WARNING):
        with tempfile.TemporaryDirectory() as tmpdir:
            working_dir = Path(tmpdir) / "agent"
            bridge_openclaw_to_copaw(
                _make_openclaw_cfg(), working_dir, profile="manager"
            )
            config = _config_json_path(working_dir)

    assert config["show_tool_details"] is True
    assert any("AGENTTEAMS_QUIET_ROOMS is deprecated" in r.message for r in caplog.records)


def test_verbose_rooms_wins_when_both_set(monkeypatch):
    monkeypatch.setenv("AGENTTEAMS_VERBOSE_ROOMS", "true")
    monkeypatch.setenv("AGENTTEAMS_QUIET_ROOMS", "true")

    with tempfile.TemporaryDirectory() as tmpdir:
        working_dir = Path(tmpdir) / "agent"
        bridge_openclaw_to_copaw(_make_openclaw_cfg(), working_dir, profile="manager")
        config = _config_json_path(working_dir)

    assert config["show_tool_details"] is True
