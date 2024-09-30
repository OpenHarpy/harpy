import duckdb

from sdk.primitives import SingletonMeta

class QuackContext(metaclass=SingletonMeta):
    def __init__(self):
        self.duck = duckdb.connect(':memory:')
    
    def restart_session(self):
        self.duck = duckdb.connect(':memory:')
    
    def sql(self, query):
        return self.duck.execute(query)
    
    def table_exists(self, table_name):
        return table_name in self.duck.table_names()
    
    def close(self):
        self.duck.close()