import unittest
from .base_test_classes import HarpyTestCase
from harpy.configs import Configs
from harpy.processing.types import OneOffClusterTask
from harpy.processing.remote_exec import RemoteExecMetadata

class FileSystem(HarpyTestCase):    
    def test_one_offs(self):
        def one_off_function() -> str:
            return RemoteExecMetadata().get_metadata('instance_id')
        
        ts = self.session.create_task_set()
        ts.add_oneoff_cluster(
            OneOffClusterTask(name='one-off-cluster-task', fun=one_off_function)
        )
        
        result = ts.run(collect=True)
        # Check if the instance IDs are different and if there are 2 of them
        self.assertEqual(len(set(result)), 2)
        self.assertEqual(len(result), 2)
        
        