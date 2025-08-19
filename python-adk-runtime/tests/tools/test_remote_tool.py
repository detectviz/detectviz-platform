import pytest
import os
import subprocess
import tempfile
import shutil
from unittest.mock import patch, MagicMock

# Mock the protobuf imports as they might not be in the test environment's path
pb_mock = MagicMock()
pbg_mock = MagicMock()
pbg_mock.ToolBridgeStub = MagicMock()

# Since we are testing RemoteTool in isolation, we mock the modules it depends on.
mock_modules = {
    'contracts.gen.python.detectviz.contracts.v1': {
        'adk_bridge_pb2': pb_mock,
        'adk_bridge_pb2_grpc': pbg_mock,
    },
    'google.adk.tools.base_tool': {
        'BaseTool': type('BaseTool', (object,), {})
    },
    'opentelemetry.propagate': MagicMock(),
    'opentelemetry.trace.propagation.tracecontext': MagicMock(),
}

# The actual module we are testing
from detectviz_adk.tools import remote_tool

@pytest.fixture(scope="module")
def certs():
    """Create temporary self-signed certificates for testing."""
    temp_dir = tempfile.mkdtemp()

    ca_key_path = os.path.join(temp_dir, "ca.key")
    ca_cert_path = os.path.join(temp_dir, "ca.crt")
    client_key_path = os.path.join(temp_dir, "client.key")
    client_csr_path = os.path.join(temp_dir, "client.csr")
    client_cert_path = os.path.join(temp_dir, "client.crt")

    try:
        # Generate CA key and cert
        subprocess.run(f"openssl genrsa -out {ca_key_path} 2048", shell=True, check=True)
        subprocess.run(f"openssl req -new -x509 -key {ca_key_path} -out {ca_cert_path} -days 365 -subj '/CN=TestCA'", shell=True, check=True)

        # Generate client key and CSR
        subprocess.run(f"openssl genrsa -out {client_key_path} 2048", shell=True, check=True)
        subprocess.run(f"openssl req -new -key {client_key_path} -out {client_csr_path} -subj '/CN=TestClient'", shell=True, check=True)

        # Sign client certificate with CA
        subprocess.run(f"openssl x509 -req -in {client_csr_path} -CA {ca_cert_path} -CAkey {ca_key_path} -CAcreateserial -out {client_cert_path} -days 365", shell=True, check=True)

        yield {
            "ca_cert": ca_cert_path,
            "client_cert": client_cert_path,
            "client_key": client_key_path,
        }
    finally:
        shutil.rmtree(temp_dir)

def test_remote_tool_insecure_connection():
    """Verify RemoteTool creates an insecure channel when TLS is disabled."""

    mock_config = {
        "grpc": {
            "listen": "localhost:12345",
            "tls": {"enabled": False}
        }
    }

    with patch('detectviz_adk.config.loader.load_config', return_value=mock_config), \
         patch('detectviz_adk.tools.remote_tool.pbg', pbg_mock), \
         patch('grpc.aio.insecure_channel') as mock_insecure_channel, \
         patch('grpc.aio.secure_channel') as mock_secure_channel:

        # Instantiate the tool, which triggers _init_channel_and_stub
        tool = remote_tool.RemoteTool(tool_id="test.tool")

        mock_insecure_channel.assert_called_once_with("localhost:12345")
        mock_secure_channel.assert_not_called()

def test_remote_tool_secure_connection_with_mtls(certs):
    """Verify RemoteTool creates a secure channel with mTLS certs."""

    mock_config = {
        "grpc": {
            "listen": "localhost:54321",
            "tls": {
                "enabled": True,
                "ca_cert": certs["ca_cert"],
                "client_cert": certs["client_cert"],
                "client_key": certs["client_key"],
            }
        }
    }

    with patch('detectviz_adk.config.loader.load_config', return_value=mock_config), \
         patch('detectviz_adk.tools.remote_tool.pbg', pbg_mock), \
         patch('grpc.aio.insecure_channel') as mock_insecure_channel, \
         patch('grpc.aio.secure_channel') as mock_secure_channel, \
         patch('grpc.ssl_channel_credentials') as mock_ssl_creds:

        # Read the cert file contents to check against later
        with open(certs['ca_cert'], 'rb') as f:
            ca_cert_bytes = f.read()
        with open(certs['client_cert'], 'rb') as f:
            client_cert_bytes = f.read()
        with open(certs['client_key'], 'rb') as f:
            client_key_bytes = f.read()

        # Instantiate the tool
        tool = remote_tool.RemoteTool(tool_id="test.secure.tool")

        # Assertions
        mock_insecure_channel.assert_not_called()
        mock_secure_channel.assert_called_once()

        # Verify that ssl_channel_credentials was called with the actual cert content
        mock_ssl_creds.assert_called_once_with(
            root_certificates=ca_cert_bytes,
            private_key=client_key_bytes,
            certificate_chain=client_cert_bytes
        )
