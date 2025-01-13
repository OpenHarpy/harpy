import unittest
from harpy.session import Session
from harpy.configs import Configs

class HarpyTestCase(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        Configs().set("harpy.tasks.node.request.count", "2")
        cls._session = Session()
        cls.__setup_env_called = False

    @classmethod
    def tearDownClass(cls):
        cls._session.close()

    def setUp(self):
        if not self.__class__.__setup_env_called:
            self.__setup_env__()
            self.__class__.__setup_env_called = True
        self.session = self._session
        
    @classmethod
    def __setup_env__(cls):
        """
        This method can be overridden to setup the environment for the test
        """
        pass
    
    def tearDown(self):
        pass