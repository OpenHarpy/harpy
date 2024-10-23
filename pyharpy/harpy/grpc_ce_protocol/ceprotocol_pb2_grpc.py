# Generated by the gRPC Python protocol compiler plugin. DO NOT EDIT!
"""Client and server classes corresponding to protobuf-defined services."""
import grpc
import warnings

from harpy.grpc_ce_protocol import ceprotocol_pb2 as ceprotocol__pb2

GRPC_GENERATED_VERSION = '1.66.1'
GRPC_VERSION = grpc.__version__
_version_not_supported = False

try:
    from grpc._utilities import first_version_is_lower
    _version_not_supported = first_version_is_lower(GRPC_VERSION, GRPC_GENERATED_VERSION)
except ImportError:
    _version_not_supported = True

if _version_not_supported:
    raise RuntimeError(
        f'The grpc package installed is at version {GRPC_VERSION},'
        + f' but the generated code in ceprotocol_pb2_grpc.py depends on'
        + f' grpcio>={GRPC_GENERATED_VERSION}.'
        + f' Please upgrade your grpc module to grpcio>={GRPC_GENERATED_VERSION}'
        + f' or downgrade your generated code using grpcio-tools<={GRPC_VERSION}.'
    )


class SessionStub(object):
    """Missing associated documentation comment in .proto file."""

    def __init__(self, channel):
        """Constructor.

        Args:
            channel: A grpc.Channel.
        """
        self.CreateSession = channel.unary_unary(
                '/proto.Session/CreateSession',
                request_serializer=ceprotocol__pb2.SessionRequest.SerializeToString,
                response_deserializer=ceprotocol__pb2.SessionHandler.FromString,
                _registered_method=True)
        self.CreateTaskSet = channel.unary_unary(
                '/proto.Session/CreateTaskSet',
                request_serializer=ceprotocol__pb2.SessionHandler.SerializeToString,
                response_deserializer=ceprotocol__pb2.TaskSetHandler.FromString,
                _registered_method=True)
        self.CloseSession = channel.unary_unary(
                '/proto.Session/CloseSession',
                request_serializer=ceprotocol__pb2.SessionHandler.SerializeToString,
                response_deserializer=ceprotocol__pb2.SessionHandler.FromString,
                _registered_method=True)
        self.GetInstanceID = channel.unary_unary(
                '/proto.Session/GetInstanceID',
                request_serializer=ceprotocol__pb2.SessionHandler.SerializeToString,
                response_deserializer=ceprotocol__pb2.InstanceMetadata.FromString,
                _registered_method=True)


class SessionServicer(object):
    """Missing associated documentation comment in .proto file."""

    def CreateSession(self, request, context):
        """Missing associated documentation comment in .proto file."""
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details('Method not implemented!')
        raise NotImplementedError('Method not implemented!')

    def CreateTaskSet(self, request, context):
        """Missing associated documentation comment in .proto file."""
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details('Method not implemented!')
        raise NotImplementedError('Method not implemented!')

    def CloseSession(self, request, context):
        """Missing associated documentation comment in .proto file."""
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details('Method not implemented!')
        raise NotImplementedError('Method not implemented!')

    def GetInstanceID(self, request, context):
        """Missing associated documentation comment in .proto file."""
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details('Method not implemented!')
        raise NotImplementedError('Method not implemented!')


def add_SessionServicer_to_server(servicer, server):
    rpc_method_handlers = {
            'CreateSession': grpc.unary_unary_rpc_method_handler(
                    servicer.CreateSession,
                    request_deserializer=ceprotocol__pb2.SessionRequest.FromString,
                    response_serializer=ceprotocol__pb2.SessionHandler.SerializeToString,
            ),
            'CreateTaskSet': grpc.unary_unary_rpc_method_handler(
                    servicer.CreateTaskSet,
                    request_deserializer=ceprotocol__pb2.SessionHandler.FromString,
                    response_serializer=ceprotocol__pb2.TaskSetHandler.SerializeToString,
            ),
            'CloseSession': grpc.unary_unary_rpc_method_handler(
                    servicer.CloseSession,
                    request_deserializer=ceprotocol__pb2.SessionHandler.FromString,
                    response_serializer=ceprotocol__pb2.SessionHandler.SerializeToString,
            ),
            'GetInstanceID': grpc.unary_unary_rpc_method_handler(
                    servicer.GetInstanceID,
                    request_deserializer=ceprotocol__pb2.SessionHandler.FromString,
                    response_serializer=ceprotocol__pb2.InstanceMetadata.SerializeToString,
            ),
    }
    generic_handler = grpc.method_handlers_generic_handler(
            'proto.Session', rpc_method_handlers)
    server.add_generic_rpc_handlers((generic_handler,))
    server.add_registered_method_handlers('proto.Session', rpc_method_handlers)


 # This class is part of an EXPERIMENTAL API.
class Session(object):
    """Missing associated documentation comment in .proto file."""

    @staticmethod
    def CreateSession(request,
            target,
            options=(),
            channel_credentials=None,
            call_credentials=None,
            insecure=False,
            compression=None,
            wait_for_ready=None,
            timeout=None,
            metadata=None):
        return grpc.experimental.unary_unary(
            request,
            target,
            '/proto.Session/CreateSession',
            ceprotocol__pb2.SessionRequest.SerializeToString,
            ceprotocol__pb2.SessionHandler.FromString,
            options,
            channel_credentials,
            insecure,
            call_credentials,
            compression,
            wait_for_ready,
            timeout,
            metadata,
            _registered_method=True)

    @staticmethod
    def CreateTaskSet(request,
            target,
            options=(),
            channel_credentials=None,
            call_credentials=None,
            insecure=False,
            compression=None,
            wait_for_ready=None,
            timeout=None,
            metadata=None):
        return grpc.experimental.unary_unary(
            request,
            target,
            '/proto.Session/CreateTaskSet',
            ceprotocol__pb2.SessionHandler.SerializeToString,
            ceprotocol__pb2.TaskSetHandler.FromString,
            options,
            channel_credentials,
            insecure,
            call_credentials,
            compression,
            wait_for_ready,
            timeout,
            metadata,
            _registered_method=True)

    @staticmethod
    def CloseSession(request,
            target,
            options=(),
            channel_credentials=None,
            call_credentials=None,
            insecure=False,
            compression=None,
            wait_for_ready=None,
            timeout=None,
            metadata=None):
        return grpc.experimental.unary_unary(
            request,
            target,
            '/proto.Session/CloseSession',
            ceprotocol__pb2.SessionHandler.SerializeToString,
            ceprotocol__pb2.SessionHandler.FromString,
            options,
            channel_credentials,
            insecure,
            call_credentials,
            compression,
            wait_for_ready,
            timeout,
            metadata,
            _registered_method=True)

    @staticmethod
    def GetInstanceID(request,
            target,
            options=(),
            channel_credentials=None,
            call_credentials=None,
            insecure=False,
            compression=None,
            wait_for_ready=None,
            timeout=None,
            metadata=None):
        return grpc.experimental.unary_unary(
            request,
            target,
            '/proto.Session/GetInstanceID',
            ceprotocol__pb2.SessionHandler.SerializeToString,
            ceprotocol__pb2.InstanceMetadata.FromString,
            options,
            channel_credentials,
            insecure,
            call_credentials,
            compression,
            wait_for_ready,
            timeout,
            metadata,
            _registered_method=True)


class TaskSetStub(object):
    """Missing associated documentation comment in .proto file."""

    def __init__(self, channel):
        """Constructor.

        Args:
            channel: A grpc.Channel.
        """
        self.DefineTask = channel.unary_unary(
                '/proto.TaskSet/DefineTask',
                request_serializer=ceprotocol__pb2.TaskDefinition.SerializeToString,
                response_deserializer=ceprotocol__pb2.TaskHandler.FromString,
                _registered_method=True)
        self.AddMap = channel.unary_unary(
                '/proto.TaskSet/AddMap',
                request_serializer=ceprotocol__pb2.MapAdder.SerializeToString,
                response_deserializer=ceprotocol__pb2.TaskAdderResult.FromString,
                _registered_method=True)
        self.AddReduce = channel.unary_unary(
                '/proto.TaskSet/AddReduce',
                request_serializer=ceprotocol__pb2.ReduceAdder.SerializeToString,
                response_deserializer=ceprotocol__pb2.TaskAdderResult.FromString,
                _registered_method=True)
        self.AddTransform = channel.unary_unary(
                '/proto.TaskSet/AddTransform',
                request_serializer=ceprotocol__pb2.TransformAdder.SerializeToString,
                response_deserializer=ceprotocol__pb2.TaskAdderResult.FromString,
                _registered_method=True)
        self.Execute = channel.unary_stream(
                '/proto.TaskSet/Execute',
                request_serializer=ceprotocol__pb2.TaskSetHandler.SerializeToString,
                response_deserializer=ceprotocol__pb2.TaskSetProgressReport.FromString,
                _registered_method=True)
        self.Dismantle = channel.unary_unary(
                '/proto.TaskSet/Dismantle',
                request_serializer=ceprotocol__pb2.TaskSetHandler.SerializeToString,
                response_deserializer=ceprotocol__pb2.TaskSetHandler.FromString,
                _registered_method=True)
        self.GetTaskSetResults = channel.unary_unary(
                '/proto.TaskSet/GetTaskSetResults',
                request_serializer=ceprotocol__pb2.TaskSetHandler.SerializeToString,
                response_deserializer=ceprotocol__pb2.TaskSetResult.FromString,
                _registered_method=True)


class TaskSetServicer(object):
    """Missing associated documentation comment in .proto file."""

    def DefineTask(self, request, context):
        """Missing associated documentation comment in .proto file."""
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details('Method not implemented!')
        raise NotImplementedError('Method not implemented!')

    def AddMap(self, request, context):
        """Missing associated documentation comment in .proto file."""
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details('Method not implemented!')
        raise NotImplementedError('Method not implemented!')

    def AddReduce(self, request, context):
        """Missing associated documentation comment in .proto file."""
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details('Method not implemented!')
        raise NotImplementedError('Method not implemented!')

    def AddTransform(self, request, context):
        """Missing associated documentation comment in .proto file."""
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details('Method not implemented!')
        raise NotImplementedError('Method not implemented!')

    def Execute(self, request, context):
        """Missing associated documentation comment in .proto file."""
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details('Method not implemented!')
        raise NotImplementedError('Method not implemented!')

    def Dismantle(self, request, context):
        """Missing associated documentation comment in .proto file."""
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details('Method not implemented!')
        raise NotImplementedError('Method not implemented!')

    def GetTaskSetResults(self, request, context):
        """Missing associated documentation comment in .proto file."""
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details('Method not implemented!')
        raise NotImplementedError('Method not implemented!')


def add_TaskSetServicer_to_server(servicer, server):
    rpc_method_handlers = {
            'DefineTask': grpc.unary_unary_rpc_method_handler(
                    servicer.DefineTask,
                    request_deserializer=ceprotocol__pb2.TaskDefinition.FromString,
                    response_serializer=ceprotocol__pb2.TaskHandler.SerializeToString,
            ),
            'AddMap': grpc.unary_unary_rpc_method_handler(
                    servicer.AddMap,
                    request_deserializer=ceprotocol__pb2.MapAdder.FromString,
                    response_serializer=ceprotocol__pb2.TaskAdderResult.SerializeToString,
            ),
            'AddReduce': grpc.unary_unary_rpc_method_handler(
                    servicer.AddReduce,
                    request_deserializer=ceprotocol__pb2.ReduceAdder.FromString,
                    response_serializer=ceprotocol__pb2.TaskAdderResult.SerializeToString,
            ),
            'AddTransform': grpc.unary_unary_rpc_method_handler(
                    servicer.AddTransform,
                    request_deserializer=ceprotocol__pb2.TransformAdder.FromString,
                    response_serializer=ceprotocol__pb2.TaskAdderResult.SerializeToString,
            ),
            'Execute': grpc.unary_stream_rpc_method_handler(
                    servicer.Execute,
                    request_deserializer=ceprotocol__pb2.TaskSetHandler.FromString,
                    response_serializer=ceprotocol__pb2.TaskSetProgressReport.SerializeToString,
            ),
            'Dismantle': grpc.unary_unary_rpc_method_handler(
                    servicer.Dismantle,
                    request_deserializer=ceprotocol__pb2.TaskSetHandler.FromString,
                    response_serializer=ceprotocol__pb2.TaskSetHandler.SerializeToString,
            ),
            'GetTaskSetResults': grpc.unary_unary_rpc_method_handler(
                    servicer.GetTaskSetResults,
                    request_deserializer=ceprotocol__pb2.TaskSetHandler.FromString,
                    response_serializer=ceprotocol__pb2.TaskSetResult.SerializeToString,
            ),
    }
    generic_handler = grpc.method_handlers_generic_handler(
            'proto.TaskSet', rpc_method_handlers)
    server.add_generic_rpc_handlers((generic_handler,))
    server.add_registered_method_handlers('proto.TaskSet', rpc_method_handlers)


 # This class is part of an EXPERIMENTAL API.
class TaskSet(object):
    """Missing associated documentation comment in .proto file."""

    @staticmethod
    def DefineTask(request,
            target,
            options=(),
            channel_credentials=None,
            call_credentials=None,
            insecure=False,
            compression=None,
            wait_for_ready=None,
            timeout=None,
            metadata=None):
        return grpc.experimental.unary_unary(
            request,
            target,
            '/proto.TaskSet/DefineTask',
            ceprotocol__pb2.TaskDefinition.SerializeToString,
            ceprotocol__pb2.TaskHandler.FromString,
            options,
            channel_credentials,
            insecure,
            call_credentials,
            compression,
            wait_for_ready,
            timeout,
            metadata,
            _registered_method=True)

    @staticmethod
    def AddMap(request,
            target,
            options=(),
            channel_credentials=None,
            call_credentials=None,
            insecure=False,
            compression=None,
            wait_for_ready=None,
            timeout=None,
            metadata=None):
        return grpc.experimental.unary_unary(
            request,
            target,
            '/proto.TaskSet/AddMap',
            ceprotocol__pb2.MapAdder.SerializeToString,
            ceprotocol__pb2.TaskAdderResult.FromString,
            options,
            channel_credentials,
            insecure,
            call_credentials,
            compression,
            wait_for_ready,
            timeout,
            metadata,
            _registered_method=True)

    @staticmethod
    def AddReduce(request,
            target,
            options=(),
            channel_credentials=None,
            call_credentials=None,
            insecure=False,
            compression=None,
            wait_for_ready=None,
            timeout=None,
            metadata=None):
        return grpc.experimental.unary_unary(
            request,
            target,
            '/proto.TaskSet/AddReduce',
            ceprotocol__pb2.ReduceAdder.SerializeToString,
            ceprotocol__pb2.TaskAdderResult.FromString,
            options,
            channel_credentials,
            insecure,
            call_credentials,
            compression,
            wait_for_ready,
            timeout,
            metadata,
            _registered_method=True)

    @staticmethod
    def AddTransform(request,
            target,
            options=(),
            channel_credentials=None,
            call_credentials=None,
            insecure=False,
            compression=None,
            wait_for_ready=None,
            timeout=None,
            metadata=None):
        return grpc.experimental.unary_unary(
            request,
            target,
            '/proto.TaskSet/AddTransform',
            ceprotocol__pb2.TransformAdder.SerializeToString,
            ceprotocol__pb2.TaskAdderResult.FromString,
            options,
            channel_credentials,
            insecure,
            call_credentials,
            compression,
            wait_for_ready,
            timeout,
            metadata,
            _registered_method=True)

    @staticmethod
    def Execute(request,
            target,
            options=(),
            channel_credentials=None,
            call_credentials=None,
            insecure=False,
            compression=None,
            wait_for_ready=None,
            timeout=None,
            metadata=None):
        return grpc.experimental.unary_stream(
            request,
            target,
            '/proto.TaskSet/Execute',
            ceprotocol__pb2.TaskSetHandler.SerializeToString,
            ceprotocol__pb2.TaskSetProgressReport.FromString,
            options,
            channel_credentials,
            insecure,
            call_credentials,
            compression,
            wait_for_ready,
            timeout,
            metadata,
            _registered_method=True)

    @staticmethod
    def Dismantle(request,
            target,
            options=(),
            channel_credentials=None,
            call_credentials=None,
            insecure=False,
            compression=None,
            wait_for_ready=None,
            timeout=None,
            metadata=None):
        return grpc.experimental.unary_unary(
            request,
            target,
            '/proto.TaskSet/Dismantle',
            ceprotocol__pb2.TaskSetHandler.SerializeToString,
            ceprotocol__pb2.TaskSetHandler.FromString,
            options,
            channel_credentials,
            insecure,
            call_credentials,
            compression,
            wait_for_ready,
            timeout,
            metadata,
            _registered_method=True)

    @staticmethod
    def GetTaskSetResults(request,
            target,
            options=(),
            channel_credentials=None,
            call_credentials=None,
            insecure=False,
            compression=None,
            wait_for_ready=None,
            timeout=None,
            metadata=None):
        return grpc.experimental.unary_unary(
            request,
            target,
            '/proto.TaskSet/GetTaskSetResults',
            ceprotocol__pb2.TaskSetHandler.SerializeToString,
            ceprotocol__pb2.TaskSetResult.FromString,
            options,
            channel_credentials,
            insecure,
            call_credentials,
            compression,
            wait_for_ready,
            timeout,
            metadata,
            _registered_method=True)


class BlockProxyStub(object):
    """Missing associated documentation comment in .proto file."""

    def __init__(self, channel):
        """Constructor.

        Args:
            channel: A grpc.Channel.
        """
        self.GetBlock = channel.unary_stream(
                '/proto.BlockProxy/GetBlock',
                request_serializer=ceprotocol__pb2.ProxyBlockHandler.SerializeToString,
                response_deserializer=ceprotocol__pb2.ProxyBlockChunk.FromString,
                _registered_method=True)
        self.PutBlock = channel.stream_unary(
                '/proto.BlockProxy/PutBlock',
                request_serializer=ceprotocol__pb2.ProxyBlockChunk.SerializeToString,
                response_deserializer=ceprotocol__pb2.ProxyBlockHandler.FromString,
                _registered_method=True)


class BlockProxyServicer(object):
    """Missing associated documentation comment in .proto file."""

    def GetBlock(self, request, context):
        """Missing associated documentation comment in .proto file."""
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details('Method not implemented!')
        raise NotImplementedError('Method not implemented!')

    def PutBlock(self, request_iterator, context):
        """Missing associated documentation comment in .proto file."""
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details('Method not implemented!')
        raise NotImplementedError('Method not implemented!')


def add_BlockProxyServicer_to_server(servicer, server):
    rpc_method_handlers = {
            'GetBlock': grpc.unary_stream_rpc_method_handler(
                    servicer.GetBlock,
                    request_deserializer=ceprotocol__pb2.ProxyBlockHandler.FromString,
                    response_serializer=ceprotocol__pb2.ProxyBlockChunk.SerializeToString,
            ),
            'PutBlock': grpc.stream_unary_rpc_method_handler(
                    servicer.PutBlock,
                    request_deserializer=ceprotocol__pb2.ProxyBlockChunk.FromString,
                    response_serializer=ceprotocol__pb2.ProxyBlockHandler.SerializeToString,
            ),
    }
    generic_handler = grpc.method_handlers_generic_handler(
            'proto.BlockProxy', rpc_method_handlers)
    server.add_generic_rpc_handlers((generic_handler,))
    server.add_registered_method_handlers('proto.BlockProxy', rpc_method_handlers)


 # This class is part of an EXPERIMENTAL API.
class BlockProxy(object):
    """Missing associated documentation comment in .proto file."""

    @staticmethod
    def GetBlock(request,
            target,
            options=(),
            channel_credentials=None,
            call_credentials=None,
            insecure=False,
            compression=None,
            wait_for_ready=None,
            timeout=None,
            metadata=None):
        return grpc.experimental.unary_stream(
            request,
            target,
            '/proto.BlockProxy/GetBlock',
            ceprotocol__pb2.ProxyBlockHandler.SerializeToString,
            ceprotocol__pb2.ProxyBlockChunk.FromString,
            options,
            channel_credentials,
            insecure,
            call_credentials,
            compression,
            wait_for_ready,
            timeout,
            metadata,
            _registered_method=True)

    @staticmethod
    def PutBlock(request_iterator,
            target,
            options=(),
            channel_credentials=None,
            call_credentials=None,
            insecure=False,
            compression=None,
            wait_for_ready=None,
            timeout=None,
            metadata=None):
        return grpc.experimental.stream_unary(
            request_iterator,
            target,
            '/proto.BlockProxy/PutBlock',
            ceprotocol__pb2.ProxyBlockChunk.SerializeToString,
            ceprotocol__pb2.ProxyBlockHandler.FromString,
            options,
            channel_credentials,
            insecure,
            call_credentials,
            compression,
            wait_for_ready,
            timeout,
            metadata,
            _registered_method=True)
