from dataclasses import dataclass
import grpc
import cloudpickle
from typing import List

from harpy.primitives import check_variable
from harpy.grpc_ce_protocol.ceprotocol_pb2 import (
    ProxyBlockChunk,
    ProxyBlockHandler,
    SessionHandler   
)
from harpy.grpc_ce_protocol.ceprotocol_pb2_grpc import (
    BlockProxyStub
)

def generate_chunks(data: bytes, session_handler: SessionHandler, chunk_size: int = 1024 * 1024) -> ProxyBlockChunk:
    for i in range(0, len(data), chunk_size):
        yield ProxyBlockChunk(SessionHandler=session_handler, BlockChunk=data[i:i + chunk_size])

class BlockReadWriteProxy:
    def __init__(self, channel: grpc.Channel, session_handler: SessionHandler):
        self._block_proxy_stub = BlockProxyStub(channel)
        self._session_handler = session_handler
    
    # rpc GetBlock (ProxyBlockHandler) returns (stream ProxyBlockChunk) {}
    def read_block(self, blockID: str) -> bytes:
        request = ProxyBlockHandler(SessionHandler=self._session_handler, BlockID=blockID)
        response = self._block_proxy_stub.GetBlock(request)
        data = b''
        for chunk in response:
            data += chunk.BlockChunk
        return data

    def write_block(self, data: bytes) -> str:
        chunk_size = 1024 * 1024  # 1MB
        result = self._block_proxy_stub.PutBlock(generate_chunks(data, self._session_handler, chunk_size))
        return result.BlockID