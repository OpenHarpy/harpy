import duckdb

from harpy.primitives import SingletonMeta

# Quack can implement a context to be used like this:
# with QuackContext() as qc:
#     qc.sql("SELECT * FROM table")
# for this to work, QuackContext must implement __enter__ and __exit__ methods
# QuackContext must also implement a close method to close the connection
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
    
    def __enter__(self):
        return self
    
    def __exit__(self, exc_type, exc_value, traceback):
        print("QuackContext exited")
        # For now we do not need to do anything here
