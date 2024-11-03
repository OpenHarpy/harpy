import unittest
from .base_test_classes import HarpyTestCase
from harpy.session import Session

class FileSystem(HarpyTestCase):
    def test_file_system_operations(self):
        self.session.fs.rm("/Volumes/data/test", recursive=True)
        self.session.fs.mkdir("/Volumes/data/test")
        return_put = self.session.fs.put("/Volumes/data/test/test.txt", "example")
        self.assertTrue(return_put)
        data_return_0 = self.session.fs.cat("/Volumes/data/test/test.txt")
        data_return_1 = self.session.fs.head("/Volumes/data/test/test.txt", 1)
        data_return_2 = self.session.fs.tail("/Volumes/data/test/test.txt", 1)
        data_return_3 = self.session.fs.get("/Volumes/data/test/test.txt")
        self.assertEqual(data_return_0, "example")
        self.assertEqual(data_return_1, "example")
        self.assertEqual(data_return_2, "example")
        self.assertEqual(data_return_3, b"example")
        self.session.fs.rm("/Volumes/data/test", recursive=True)
        