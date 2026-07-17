import asyncio
import json
import sys
import types

import pytest

from copaw_worker.config import WorkerConfig
from copaw_worker.health import HealthState, check_matrix_service
from copaw_worker.worker import Worker


class _FakeWorkerAPIServer:
    instances = []

    def __init__(self, *, host, port, liveness_handler, readiness_handler):
        self.host = host
        self.port = port
        self.liveness_handler = liveness_handler
        self.readiness_handler = readiness_handler
        self.started = False
        self.stopped = False
        self.__class__.instances.append(self)

    async def start(self):
        self.started = True

    async def stop(self):
        self.stopped = True


@pytest.fixture(autouse=True)
def fake_worker_api(monkeypatch):
    _FakeWorkerAPIServer.instances = []
    monkeypatch.setattr("copaw_worker.worker.WorkerAPIServer", _FakeWorkerAPIServer)
    return _FakeWorkerAPIServer


@pytest.fixture
def anyio_backend():
    return "asyncio"


def _config(tmp_path, **overrides):
    base = dict(
        worker_name="alice",
        minio_endpoint="http://minio:9000",
        minio_access_key="minio",
        minio_secret_key="password",
        install_dir=tmp_path,
    )
    base.update(overrides)
    return WorkerConfig(**base)


async def _finished_push_loop(*_args, **_kwargs):
    return None


async def _finished_sync_loop(*_args, **_kwargs):
    return None


def test_worker_port_defaults_to_console_port_plus_one(tmp_path):
    config = _config(tmp_path, console_port=18088)
    assert config.worker_port == 18089


def test_worker_port_can_be_explicit(tmp_path):
    config = _config(tmp_path, console_port=18088, worker_port=19090)
    assert config.worker_port == 19090


def test_join_pending_matrix_invites_accepts_invited_rooms(tmp_path, monkeypatch):
    import urllib.request

    requests = []

    class FakeResponse:
        def __init__(self, data=None):
            self._data = data or {}

        def __enter__(self):
            return self

        def __exit__(self, *_):
            return False

        def read(self):
            return json.dumps(self._data).encode()

    def fake_urlopen(req, timeout=None):
        requests.append((req.get_method(), req.full_url))
        if "/sync?" in req.full_url:
            return FakeResponse({
                "rooms": {
                    "invite": {
                        "!team:test": {},
                        "!dm:test": {},
                    }
                }
            })
        return FakeResponse()

    monkeypatch.setattr(urllib.request, "urlopen", fake_urlopen)

    worker = Worker(_config(tmp_path))
    worker._join_pending_matrix_invites({
        "channels": {
            "matrix": {
                "homeserver": "http://matrix:6167",
                "accessToken": "tok",
            }
        }
    })

    assert requests[0] == (
        "GET",
        "http://matrix:6167/_matrix/client/v3/sync?timeout=0&full_state=true",
    )
    assert requests[1] == (
        "POST",
        "http://matrix:6167/_matrix/client/v3/join/%21team%3Atest",
    )
    assert requests[2] == (
        "POST",
        "http://matrix:6167/_matrix/client/v3/join/%21dm%3Atest",
    )


@pytest.mark.anyio
async def test_worker_start_succeeds_on_happy_path(tmp_path, monkeypatch):
    monkeypatch.setattr(Worker, "_ensure_mc", lambda _self: None)
    monkeypatch.setattr("copaw_worker.sync.FileSync.mirror_all", lambda _self: None)
    monkeypatch.setattr(
        "copaw_worker.sync.FileSync.get_config",
        lambda _self: {"channels": {"matrix": {}}},
    )
    monkeypatch.setattr("copaw_worker.sync.FileSync.list_skills", lambda _self: [])
    monkeypatch.setattr(Worker, "_matrix_relogin", lambda _self, cfg: cfg)
    monkeypatch.setattr("copaw_worker.worker.bridge_openclaw_to_copaw", lambda *_a, **_k: None)
    monkeypatch.setattr("copaw_worker.worker.push_loop", _finished_push_loop)
    monkeypatch.setattr("copaw_worker.worker.sync_loop", _finished_sync_loop)

    worker = Worker(_config(tmp_path))
    assert await worker.start() is True
    await worker.stop()


@pytest.mark.anyio
async def test_worker_start_fails_when_startup_mirror_fails(tmp_path, monkeypatch):
    async def instant_sleep(_seconds):
        return None

    def fail_mirror(_self):
        raise RuntimeError("minio unavailable")

    monkeypatch.setattr(Worker, "_ensure_mc", lambda _self: None)
    monkeypatch.setattr("copaw_worker.sync.FileSync.mirror_all", fail_mirror)
    monkeypatch.setattr("copaw_worker.worker.asyncio.sleep", instant_sleep)

    worker = Worker(_config(tmp_path))
    assert await worker.start() is False


@pytest.mark.anyio
async def test_worker_start_fails_when_bridge_fails(tmp_path, monkeypatch):
    def fail_bridge(*_args, **_kwargs):
        raise ValueError("bad openclaw config")

    monkeypatch.setattr(Worker, "_ensure_mc", lambda _self: None)
    monkeypatch.setattr("copaw_worker.sync.FileSync.mirror_all", lambda _self: None)
    monkeypatch.setattr(
        "copaw_worker.sync.FileSync.get_config",
        lambda _self: {"channels": {"matrix": {}}},
    )
    monkeypatch.setattr("copaw_worker.sync.FileSync.list_skills", lambda _self: [])
    monkeypatch.setattr(Worker, "_matrix_relogin", lambda _self, cfg: cfg)
    monkeypatch.setattr("copaw_worker.worker.bridge_openclaw_to_copaw", fail_bridge)

    worker = Worker(_config(tmp_path))
    assert await worker.start() is False


@pytest.mark.anyio
async def test_health_state_snapshot_marks_unhealthy_component(tmp_path):
    health = HealthState(tmp_path / "alice" / ".copaw" / "health.json")
    health.update("model", "unhealthy", "model provider is unreachable")
    snapshot = health.snapshot()
    assert snapshot.healthiness == "unhealthy"
    assert snapshot.components["model"].message == "model provider is unreachable"


def test_check_matrix_service_reports_reachable_homeserver(monkeypatch):
    import urllib.request

    class FakeResponse:
        def __init__(self):
            self.status = 200

        def __enter__(self):
            return self

        def __exit__(self, *_):
            return False

    monkeypatch.setattr(urllib.request, "urlopen", lambda *_a, **_k: FakeResponse())

    result = check_matrix_service("http://matrix:6167", timeout=1)
    assert result.healthiness == "healthy"
    assert "reachable" in result.message


@pytest.mark.anyio
async def test_worker_installs_hooks_before_copaw_runner(tmp_path, monkeypatch):
    calls = []

    fake_hooks = types.ModuleType("copaw_worker.hooks")
    fake_hooks.install_tool_hooks = lambda: calls.append("hooks")
    monkeypatch.setitem(sys.modules, "copaw_worker.hooks", fake_hooks)

    config = _config(tmp_path, console_port=8088, worker_port=8089)
    worker = Worker(config)
    worker.config.console_port = 0

    async def fake_headless():
        calls.append("headless")

    monkeypatch.setattr(worker, "_run_copaw_headless", fake_headless)

    await worker._run_copaw()

    assert calls == ["hooks", "headless"]
