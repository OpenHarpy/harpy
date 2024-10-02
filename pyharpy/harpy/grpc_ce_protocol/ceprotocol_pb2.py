# -*- coding: utf-8 -*-
# Generated by the protocol buffer compiler.  DO NOT EDIT!
# NO CHECKED-IN PROTOBUF GENCODE
# source: ceprotocol.proto
# Protobuf Python Version: 5.27.2
"""Generated protocol buffer code."""
from google.protobuf import descriptor as _descriptor
from google.protobuf import descriptor_pool as _descriptor_pool
from google.protobuf import runtime_version as _runtime_version
from google.protobuf import symbol_database as _symbol_database
from google.protobuf.internal import builder as _builder
_runtime_version.ValidateProtobufRuntimeVersion(
    _runtime_version.Domain.PUBLIC,
    5,
    27,
    2,
    '',
    'ceprotocol.proto'
)
# @@protoc_insertion_point(imports)

_sym_db = _symbol_database.Default()




DESCRIPTOR = _descriptor_pool.Default().AddSerializedFile(b'\n\x10\x63\x65protocol.proto\x12\x05proto\"u\n\x0eSessionRequest\x12\x33\n\x07Options\x18\x01 \x03(\x0b\x32\".proto.SessionRequest.OptionsEntry\x1a.\n\x0cOptionsEntry\x12\x0b\n\x03key\x18\x01 \x01(\t\x12\r\n\x05value\x18\x02 \x01(\t:\x02\x38\x01\"4\n\x0eSessionHandler\x12\x11\n\tsessionId\x18\x01 \x01(\t\x12\x0f\n\x07Success\x18\x02 \x01(\x08\"\xb6\x02\n\x13TaskDefinitionChunk\x12\x0c\n\x04Name\x18\x01 \x01(\t\x12\x16\n\x0e\x43\x61llableBinary\x18\x02 \x01(\x0c\x12H\n\x0f\x41rgumentsBinary\x18\x03 \x03(\x0b\x32/.proto.TaskDefinitionChunk.ArgumentsBinaryEntry\x12\x42\n\x0cKwargsBinary\x18\x04 \x03(\x0b\x32,.proto.TaskDefinitionChunk.KwargsBinaryEntry\x1a\x36\n\x14\x41rgumentsBinaryEntry\x12\x0b\n\x03key\x18\x01 \x01(\r\x12\r\n\x05value\x18\x02 \x01(\x0c:\x02\x38\x01\x1a\x33\n\x11KwargsBinaryEntry\x12\x0b\n\x03key\x18\x01 \x01(\t\x12\r\n\x05value\x18\x02 \x01(\x0c:\x02\x38\x01\"\x1d\n\x0bTaskHandler\x12\x0e\n\x06TaskID\x18\x01 \x01(\t\"4\n\x0eTaskSetHandler\x12\x11\n\ttaskSetId\x18\x01 \x01(\t\x12\x0f\n\x07Success\x18\x02 \x01(\x08\"9\n\rTaskSetStatus\x12\x11\n\tTaskSetID\x18\x01 \x01(\t\x12\x15\n\rTaskSetStatus\x18\x02 \x01(\t\"\x8f\x01\n\x12TaskSetResultChunk\x12\x11\n\tTaskRunID\x18\x01 \x01(\t\x12\x1f\n\x17ObjectReturnBinaryChunk\x18\x02 \x01(\x0c\x12\x19\n\x11StdoutBinaryChunk\x18\x03 \x01(\x0c\x12\x19\n\x11StderrBinaryChunk\x18\x04 \x01(\x0c\x12\x0f\n\x07Success\x18\x05 \x01(\x08\"9\n\x0fTaskAdderResult\x12\x0f\n\x07Success\x18\x01 \x01(\x08\x12\x15\n\rErrorMesssage\x18\x02 \x01(\t\"h\n\x08MapAdder\x12-\n\x0etaskSetHandler\x18\x01 \x01(\x0b\x32\x15.proto.TaskSetHandler\x12-\n\x11MappersDefinition\x18\x02 \x03(\x0b\x32\x12.proto.TaskHandler\"k\n\x0bReduceAdder\x12-\n\x0etaskSetHandler\x18\x01 \x01(\x0b\x32\x15.proto.TaskSetHandler\x12-\n\x11ReducerDefinition\x18\x02 \x01(\x0b\x32\x12.proto.TaskHandler\"r\n\x0eTransformAdder\x12-\n\x0etaskSetHandler\x18\x01 \x01(\x0b\x32\x15.proto.TaskSetHandler\x12\x31\n\x15TransformerDefinition\x18\x02 \x01(\x0b\x32\x12.proto.TaskHandler\"\xb8\x02\n\x15TaskSetProgressReport\x12-\n\x0etaskSetHandler\x18\x01 \x01(\x0b\x32\x15.proto.TaskSetHandler\x12)\n\x0cProgressType\x18\x02 \x01(\x0e\x32\x13.proto.ProgressType\x12\x11\n\tRelatedID\x18\x03 \x01(\t\x12\x17\n\x0fProgressMessage\x18\x04 \x01(\t\x12\x15\n\rStatusMessage\x18\x05 \x01(\t\x12J\n\x0fProgressDetails\x18\x06 \x03(\x0b\x32\x31.proto.TaskSetProgressReport.ProgressDetailsEntry\x1a\x36\n\x14ProgressDetailsEntry\x12\x0b\n\x03key\x18\x01 \x01(\t\x12\r\n\x05value\x18\x02 \x01(\t:\x02\x38\x01*L\n\x0cProgressType\x12\x13\n\x0fTaskSetProgress\x10\x00\x12\x15\n\x11TaskGroupProgress\x10\x01\x12\x10\n\x0cTaskProgress\x10\x02\x32\xcb\x01\n\x07Session\x12?\n\rCreateSession\x12\x15.proto.SessionRequest\x1a\x15.proto.SessionHandler\"\x00\x12?\n\rCreateTaskSet\x12\x15.proto.SessionHandler\x1a\x15.proto.TaskSetHandler\"\x00\x12>\n\x0c\x43loseSession\x12\x15.proto.SessionHandler\x1a\x15.proto.SessionHandler\"\x00\x32\xc8\x03\n\x07TaskSet\x12@\n\nDefineTask\x12\x1a.proto.TaskDefinitionChunk\x1a\x12.proto.TaskHandler\"\x00(\x01\x12\x33\n\x06\x41\x64\x64Map\x12\x0f.proto.MapAdder\x1a\x16.proto.TaskAdderResult\"\x00\x12\x39\n\tAddReduce\x12\x12.proto.ReduceAdder\x1a\x16.proto.TaskAdderResult\"\x00\x12?\n\x0c\x41\x64\x64Transform\x12\x15.proto.TransformAdder\x1a\x16.proto.TaskAdderResult\"\x00\x12\x42\n\x07\x45xecute\x12\x15.proto.TaskSetHandler\x1a\x1c.proto.TaskSetProgressReport\"\x00\x30\x01\x12;\n\tDismantle\x12\x15.proto.TaskSetHandler\x1a\x15.proto.TaskSetHandler\"\x00\x12I\n\x11GetTaskSetResults\x12\x15.proto.TaskSetHandler\x1a\x19.proto.TaskSetResultChunk\"\x00\x30\x01\x62\x06proto3')

_globals = globals()
_builder.BuildMessageAndEnumDescriptors(DESCRIPTOR, _globals)
_builder.BuildTopDescriptorsAndMessages(DESCRIPTOR, 'ceprotocol_pb2', _globals)
if not _descriptor._USE_C_DESCRIPTORS:
  DESCRIPTOR._loaded_options = None
  _globals['_SESSIONREQUEST_OPTIONSENTRY']._loaded_options = None
  _globals['_SESSIONREQUEST_OPTIONSENTRY']._serialized_options = b'8\001'
  _globals['_TASKDEFINITIONCHUNK_ARGUMENTSBINARYENTRY']._loaded_options = None
  _globals['_TASKDEFINITIONCHUNK_ARGUMENTSBINARYENTRY']._serialized_options = b'8\001'
  _globals['_TASKDEFINITIONCHUNK_KWARGSBINARYENTRY']._loaded_options = None
  _globals['_TASKDEFINITIONCHUNK_KWARGSBINARYENTRY']._serialized_options = b'8\001'
  _globals['_TASKSETPROGRESSREPORT_PROGRESSDETAILSENTRY']._loaded_options = None
  _globals['_TASKSETPROGRESSREPORT_PROGRESSDETAILSENTRY']._serialized_options = b'8\001'
  _globals['_PROGRESSTYPE']._serialized_start=1508
  _globals['_PROGRESSTYPE']._serialized_end=1584
  _globals['_SESSIONREQUEST']._serialized_start=27
  _globals['_SESSIONREQUEST']._serialized_end=144
  _globals['_SESSIONREQUEST_OPTIONSENTRY']._serialized_start=98
  _globals['_SESSIONREQUEST_OPTIONSENTRY']._serialized_end=144
  _globals['_SESSIONHANDLER']._serialized_start=146
  _globals['_SESSIONHANDLER']._serialized_end=198
  _globals['_TASKDEFINITIONCHUNK']._serialized_start=201
  _globals['_TASKDEFINITIONCHUNK']._serialized_end=511
  _globals['_TASKDEFINITIONCHUNK_ARGUMENTSBINARYENTRY']._serialized_start=404
  _globals['_TASKDEFINITIONCHUNK_ARGUMENTSBINARYENTRY']._serialized_end=458
  _globals['_TASKDEFINITIONCHUNK_KWARGSBINARYENTRY']._serialized_start=460
  _globals['_TASKDEFINITIONCHUNK_KWARGSBINARYENTRY']._serialized_end=511
  _globals['_TASKHANDLER']._serialized_start=513
  _globals['_TASKHANDLER']._serialized_end=542
  _globals['_TASKSETHANDLER']._serialized_start=544
  _globals['_TASKSETHANDLER']._serialized_end=596
  _globals['_TASKSETSTATUS']._serialized_start=598
  _globals['_TASKSETSTATUS']._serialized_end=655
  _globals['_TASKSETRESULTCHUNK']._serialized_start=658
  _globals['_TASKSETRESULTCHUNK']._serialized_end=801
  _globals['_TASKADDERRESULT']._serialized_start=803
  _globals['_TASKADDERRESULT']._serialized_end=860
  _globals['_MAPADDER']._serialized_start=862
  _globals['_MAPADDER']._serialized_end=966
  _globals['_REDUCEADDER']._serialized_start=968
  _globals['_REDUCEADDER']._serialized_end=1075
  _globals['_TRANSFORMADDER']._serialized_start=1077
  _globals['_TRANSFORMADDER']._serialized_end=1191
  _globals['_TASKSETPROGRESSREPORT']._serialized_start=1194
  _globals['_TASKSETPROGRESSREPORT']._serialized_end=1506
  _globals['_TASKSETPROGRESSREPORT_PROGRESSDETAILSENTRY']._serialized_start=1452
  _globals['_TASKSETPROGRESSREPORT_PROGRESSDETAILSENTRY']._serialized_end=1506
  _globals['_SESSION']._serialized_start=1587
  _globals['_SESSION']._serialized_end=1790
  _globals['_TASKSET']._serialized_start=1793
  _globals['_TASKSET']._serialized_end=2249
# @@protoc_insertion_point(module_scope)