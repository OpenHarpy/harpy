from harpy.primitives import SingletonMeta

class Configs(metaclass=SingletonMeta):
    def __init__(self):
        self.configs = {}

    def get(self, key):
        if key not in self.configs:
            return None
        return self.configs[key]
    
    def set(self, key, value):
        self.configs[key] = value
        
    def getAllConfigs(self):
        return self.configs