from harpy.primitives import SingletonMeta
class Configs(metaclass=SingletonMeta):
    def __init__(self):
        self.configs = {}

    def get(self, key):
        return self.configs[key]
    
    def set(self, key, value):
        self.configs[key] = value
        
    def getAllConfigs(self):
        return self.configs