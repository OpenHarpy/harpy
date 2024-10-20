"""
harpy.tasksets.tasks module
---------------------------
This module contains a collection of functions "tasks" that can be used to accomplish a specific goal.
"""
from harpy.processing.types import MapTask, ReduceTask, TransformTask
from harpy.session import Session
from harpy.tasksets import TaskSet
from harpy.quack import QuackContext
import harpy.tasksets.task_definitions as task_definitions 
import pandas as pd
import pyarrow as pa

# Taskset functions
def taskset_from_sql(sql:str, engine:str = 'pyarrow') -> TaskSet:
    session = Session().create_session()
    ts = session.create_task_set()
    if engine == 'pyarrow':
        ts.add_maps([MapTask(name="duckdb-query-arrow", fun=task_definitions.definition_quack_query_arrow_table, args=[], kwargs={"query": sql, "rows_per_batch": 1000000})])
    elif engine == 'pandas':
        ts.add_maps([MapTask(name="duckdb-query-pandas", fun=task_definitions.definition_quack_query_pandas, args=[], kwargs={"query": sql})])
    else:
        raise NotImplementedError(f"Return type {engine} is not implemented")
    return ts

def write_to_deltalake(ts:TaskSet, path:str, max_rows_per_file:int=1000000, max_partitions:int=4, engine:str='pyarrow') -> TaskSet:
    ts.add_transform(
        TransformTask(
            name="write-delta-lake",
            fun=task_definitions.definition_write_delta_lake,
            args=[],
            kwargs={"path": path, "max_rows_per_file": max_rows_per_file, "max_partitions": max_partitions, "engine": engine}
        )
    )
    return ts
    