"""
harpy.tasksets.tasks module
---------------------------
This module contains a collection of functions "tasks" that can be used to accomplish a specific goal.
"""
from harpy.quack import QuackContext
import pandas as pd
import pyarrow as pa
from deltalake import write_deltalake

# SQL functions 
def definition_quack_query_pandas(query: str) -> pd.DataFrame:
    with QuackContext() as qc:
        return qc.sql(query).fetchdf()

def definition_quack_query_arrow_table(query: str, rows_per_batch:int) -> pa.Table:
    with QuackContext() as qc:
        return qc.sql(query).arrow(rows_per_batch=rows_per_batch)

def definition_write_delta_lake(table: pa.Table, path: str, max_rows_per_file:int, max_partitions:int, engine:str) -> None:
    write_deltalake(path, table, max_rows_per_file=max_rows_per_file, max_partitions=max_partitions, engine=engine)
