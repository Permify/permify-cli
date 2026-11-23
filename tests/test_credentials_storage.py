"""
Test suite for Permify CLI credential and TLS storage resolution.
Issue: Permify/permify-cli#2 - Endpoint, Token, Cert Path, and Cert Key Storage for CLI Tool.
"""

import os
import stat
import tempfile
import yaml
import pytest


class ConfigManager:
    """Python simulation model representing core/config/config.go behavior."""

    def __init__(self, file_path=None, profile="default"):
        self.file_path = file_path
        self.profile = profile
        self.configs = {}

    def new(self, file_path, profile, initial_config=None):
        self.file_path = file_path
        self.profile = profile
        self.configs = {profile: initial_config or {}}
        self.write()

    def load(self, file_path, profile="default"):
        if not os.path.exists(file_path):
            raise FileNotFoundError(f"config file {file_path} does not exist")
        with open(file_path, "r", encoding="utf-8") as f:
            data = yaml.safe_load(f) or {}
        self.file_path = file_path
        self.profile = profile
        self.configs = data
        return self.get_profile_config(profile)

    def write(self):
        if not self.file_path:
            raise ValueError("file path not set")
        data = yaml.dump(self.configs, default_flow_style=False)
        # Emulate 0600 secure file write
        with open(self.file_path, "w", encoding="utf-8") as f:
            f.write(data)
        try:
            os.chmod(self.file_path, stat.S_IRUSR | stat.S_IWUSR)
        except OSError:
            pass

    def get_profile_config(self, profile=None):
        prof = profile or self.profile
        if prof not in self.configs:
            raise KeyError(f"profile {prof} not found in config")
        cfg = dict(self.configs[prof])
        # Resolve endpoint / permify_url compatibility
        if not cfg.get("endpoint") and cfg.get("permify_url"):
            cfg["endpoint"] = cfg["permify_url"]
        elif not cfg.get("permify_url") and cfg.get("endpoint"):
            cfg["permify_url"] = cfg["endpoint"]
        return cfg

    def set_profile_config(self, profile, config_dict):
        self.configs[profile] = config_dict


class CredentialResolver:
    """Python simulation model representing core/client/grpc.go resolution logic."""

    @staticmethod
    def resolve_credentials(cli_endpoint=None, config=None, env=None):
        env = env or {}
        config = config or {}

        # 1. Resolve endpoint
        endpoint = cli_endpoint or config.get("endpoint") or config.get("permify_url") or env.get("PERMIFY_ENDPOINT")
        if not endpoint:
            raise ValueError("permify endpoint is required: configure via `permctl configure` or PERMIFY_ENDPOINT")

        # 2. Resolve token
        token = config.get("token") or env.get("PERMIFY_TOKEN") or ""

        # 3. Resolve cert_path & cert_key
        cert_path = config.get("cert_path") or env.get("PERMIFY_CERT_PATH") or ""
        cert_key = config.get("cert_key") or env.get("PERMIFY_CERT_KEY") or ""

        # 4. Resolve insecure mode
        insecure = config.get("insecure", False)
        if env.get("PERMIFY_INSECURE") in ("true", "1", True):
            insecure = True

        # Validation checks
        if cert_key and not cert_path:
            raise ValueError("cert_key provided without cert_path: both must be supplied for client TLS")

        if cert_path and not os.path.exists(cert_path):
            raise FileNotFoundError(f"certificate file not found: {cert_path}")

        if cert_key and not os.path.exists(cert_key):
            raise FileNotFoundError(f"certificate key file not found: {cert_key}")

        # Normalization
        is_https = endpoint.lower().startswith("https://")
        clean_endpoint = endpoint
        for prefix in ("https://", "http://"):
            if clean_endpoint.lower().startswith(prefix):
                clean_endpoint = clean_endpoint[len(prefix):]
        clean_endpoint = clean_endpoint.rstrip("/")

        use_tls = (is_https or bool(cert_path) or bool(cert_key) or config.get("ssl_enabled", False)) and not insecure

        auth_header = f"Bearer {token}" if token else None

        return {
            "endpoint": clean_endpoint,
            "token": token,
            "cert_path": cert_path,
            "cert_key": cert_key,
            "use_tls": use_tls,
            "insecure": insecure,
            "auth_header": auth_header,
        }


# =========================================================================
# Unit Tests
# =========================================================================

def test_yaml_storage_and_retrieval(tmp_path):
    """Test full round-trip serialization of endpoint, token, cert_path and cert_key."""
    cfg_file = tmp_path / "permctl.yaml"
    manager = ConfigManager()

    initial = {
        "endpoint": "api.permify.cloud:443",
        "tenant": "t_production_001",
        "token": "secret_permify_bearer_token_xyz987",
        "cert_path": str(tmp_path / "client.crt"),
        "cert_key": str(tmp_path / "client.key"),
        "insecure": False,
    }

    # Create dummy cert/key files
    (tmp_path / "client.crt").write_text("-----BEGIN CERTIFICATE-----\n...\n-----END CERTIFICATE-----")
    (tmp_path / "client.key").write_text("-----BEGIN PRIVATE KEY-----\n...\n-----END PRIVATE KEY-----")

    manager.new(str(cfg_file), "production", initial)

    # Read back and verify
    loaded = manager.load(str(cfg_file), "production")
    assert loaded["endpoint"] == "api.permify.cloud:443"
    assert loaded["tenant"] == "t_production_001"
    assert loaded["token"] == "secret_permify_bearer_token_xyz987"
    assert loaded["cert_path"] == str(tmp_path / "client.crt")
    assert loaded["cert_key"] == str(tmp_path / "client.key")
    assert loaded["insecure"] is False


def test_permission_bits_on_config_file(tmp_path):
    """Verify that credentials file is written with safe permissions (0600 on POSIX)."""
    cfg_file = tmp_path / "secure_permctl.yaml"
    manager = ConfigManager()
    manager.new(str(cfg_file), "default", {
        "endpoint": "localhost:3478",
        "token": "sensitive_jwt_token",
    })

    assert cfg_file.exists()
    if os.name != "nt":
        file_mode = stat.S_IMODE(os.stat(str(cfg_file)).st_mode)
        assert file_mode & stat.S_IRWXG == 0, "Group permissions should be 0"
        assert file_mode & stat.S_IRWXO == 0, "Other permissions should be 0"


def test_multi_profile_isolation(tmp_path):
    """Verify credentials remain isolated across different configuration profiles."""
    cfg_file = tmp_path / "multi_profile.yaml"
    manager = ConfigManager()

    manager.new(str(cfg_file), "default", {
        "endpoint": "localhost:3478",
        "tenant": "t_dev",
        "token": "dev_token",
    })

    # Add staging and production profiles
    manager.set_profile_config("staging", {
        "endpoint": "staging.permify.io:443",
        "tenant": "t_staging",
        "token": "staging_token_456",
    })
    manager.set_profile_config("production", {
        "endpoint": "prod.permify.io:443",
        "tenant": "t_prod",
        "token": "prod_token_789",
        "cert_path": "/etc/ssl/prod.crt",
        "cert_key": "/etc/ssl/prod.key",
    })
    manager.write()

    # Verify each profile reads its own credentials
    dev_cfg = manager.load(str(cfg_file), "default")
    assert dev_cfg["token"] == "dev_token"

    staging_cfg = manager.load(str(cfg_file), "staging")
    assert staging_cfg["token"] == "staging_token_456"

    prod_cfg = manager.load(str(cfg_file), "production")
    assert prod_cfg["token"] == "prod_token_789"
    assert prod_cfg["cert_path"] == "/etc/ssl/prod.crt"


def test_endpoint_resolution_hierarchy():
    """Verify resolution hierarchy: CLI argument > Config file > Environment Variable."""
    # 1. CLI argument takes precedence over everything
    res = CredentialResolver.resolve_credentials(
        cli_endpoint="override.host:3478",
        config={"endpoint": "config.host:3478"},
        env={"PERMIFY_ENDPOINT": "env.host:3478"},
    )
    assert res["endpoint"] == "override.host:3478"

    # 2. Config file takes precedence over env var
    res2 = CredentialResolver.resolve_credentials(
        cli_endpoint=None,
        config={"endpoint": "config.host:3478"},
        env={"PERMIFY_ENDPOINT": "env.host:3478"},
    )
    assert res2["endpoint"] == "config.host:3478"

    # 3. Env var used when config and cli are empty
    res3 = CredentialResolver.resolve_credentials(
        cli_endpoint=None,
        config={},
        env={"PERMIFY_ENDPOINT": "env.host:3478"},
    )
    assert res3["endpoint"] == "env.host:3478"

    # 4. Error raised when no endpoint is available
    with pytest.raises(ValueError, match="permify endpoint is required"):
        CredentialResolver.resolve_credentials(cli_endpoint=None, config={}, env={})


def test_token_and_header_generation():
    """Verify Bearer token parsing and gRPC metadata authorization header generation."""
    res = CredentialResolver.resolve_credentials(
        cli_endpoint="localhost:3478",
        config={"token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.test"},
    )
    assert res["token"] == "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.test"
    assert res["auth_header"] == "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.test"

    # Env var fallback for token
    res_env = CredentialResolver.resolve_credentials(
        cli_endpoint="localhost:3478",
        config={},
        env={"PERMIFY_TOKEN": "token_from_env_123"},
    )
    assert res_env["token"] == "token_from_env_123"
    assert res_env["auth_header"] == "Bearer token_from_env_123"


def test_tls_and_cert_key_validation(tmp_path):
    """Verify TLS certificate and key validation rules."""
    cert_file = tmp_path / "valid.crt"
    key_file = tmp_path / "valid.key"
    cert_file.write_text("CERT")
    key_file.write_text("KEY")

    # 1. Valid mTLS configuration
    res = CredentialResolver.resolve_credentials(
        cli_endpoint="localhost:3478",
        config={"cert_path": str(cert_file), "cert_key": str(key_file)},
    )
    assert res["use_tls"] is True
    assert res["cert_path"] == str(cert_file)
    assert res["cert_key"] == str(key_file)

    # 2. Key provided without cert path -> Error
    with pytest.raises(ValueError, match="cert_key provided without cert_path"):
        CredentialResolver.resolve_credentials(
            cli_endpoint="localhost:3478",
            config={"cert_key": str(key_file)},
        )

    # 3. Non-existent cert path -> Error
    with pytest.raises(FileNotFoundError, match="certificate file not found"):
        CredentialResolver.resolve_credentials(
            cli_endpoint="localhost:3478",
            config={"cert_path": str(tmp_path / "non_existent.crt")},
        )

    # 4. Insecure mode overrides TLS
    res_insecure = CredentialResolver.resolve_credentials(
        cli_endpoint="https://localhost:3478",
        config={"cert_path": str(cert_file), "insecure": True},
    )
    assert res_insecure["use_tls"] is False
    assert res_insecure["insecure"] is True


def test_url_prefix_normalization():
    """Verify endpoint strings with https://, http://, or trailing slashes are properly normalized."""
    test_cases = [
        ("http://127.0.0.1:3478/", "127.0.0.1:3478"),
        ("https://permify.internal:8443/", "permify.internal:8443"),
        ("localhost:3478", "localhost:3478"),
    ]
    for raw, expected in test_cases:
        res = CredentialResolver.resolve_credentials(cli_endpoint=raw)
        assert res["endpoint"] == expected
