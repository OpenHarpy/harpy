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
from uuid import uuid4

# Taskset functions
def taskset_from_sql(sql:str, engine:str = 'pyarrow') -> TaskSet:
    session = Session()
    ts = session.create_task_set()
    if engine == 'pyarrow':
        ts.add_maps([MapTask(name="duckdb-query-arrow", fun=task_definitions.definition_quack_query_arrow_table, args=[], kwargs={"query": sql, "rows_per_batch": 1000000})])
    elif engine == 'pandas':
        ts.add_maps([MapTask(name="duckdb-query-pandas", fun=task_definitions.definition_quack_query_pandas, args=[], kwargs={"query": sql})])
    else:
        raise NotImplementedError(f"Return type {engine} is not implemented")
    return ts

def rewrite_repartition(source_type, source_location, destition_location, target_rows_per_partition=100000, mode="overwrite"):
    # Session
    session = Session()
    # Build the from clause
    if source_type == "delta":
        from_clause = f"delta_scan('{source_location}')"
    elif source_type == "parquet":
        from_clause = f"read_parquet('{source_location}')"
    else:
        raise ValueError(f"Unknown source type: {source_type}")

    sql = f"""
    WITH count as (
        SELECT count(*) AS count 
        FROM {from_clause}
    ),
    partitions as (
        SELECT 
            CAST(ceil(count / {target_rows_per_partition}) AS INTEGER) as partitions, count
        FROM count
    ),
    partition_rows as (
        SELECT 
            unnest(GENERATE_SERIES(1, partitions)) as partition_index,
            count // partitions as rows_per_partition,
            count % partitions as remainder,
            partitions as max_partitions,
            count
        FROM partitions
    )
    SELECT
        partition_index,
        -- Create the limit clause for the partition
        CASE
            WHEN partition_index = max_partitions THEN rows_per_partition + remainder
            ELSE rows_per_partition
        END AS limit,
        -- Create the offset clause for the partition
        CASE
            WHEN partition_index = 1 THEN 0
            WHEN partition_index = max_partitions THEN (partition_index - 1) * rows_per_partition + remainder
            ELSE (partition_index - 1) * rows_per_partition
        END AS offset,
        count
    FROM partition_rows
    """

    ts = session.create_task_set()
    ts.add_maps([
        MapTask(
            'partition-plan', 
            task_definitions.definition_quack_query_pandas,
            [],
            {'query': sql}
        )
    ])
    results = ts.run()
    df = results[0]

    # Mapper tasks for each partition
    # Write UUID to avoid conflicts
    uuid = str(uuid4())
    ts = session.create_task_set()
    maps = []
    for _, row in df.iterrows():
        sql = f"""
            COPY (
                SELECT * FROM {from_clause} LIMIT {row.limit} OFFSET {row.offset}
            ) TO '{destition_location}/{uuid}_{row.partition_index}.parquet' (FORMAT PARQUET, ROW_GROUP_SIZE 1024, COMPRESSION SNAPPY)
        """
        maps.append(
            MapTask(
                f'partition-{row["partition_index"]}', 
                task_definitions.definition_quack_query_pandas,
                [],
                {'query': sql}
            )
        )
    ts.add_maps(maps)
    if mode == "overwrite":
        session.fs.rm(destition_location, recursive=True)
    elif mode == "append":
        pass
    else:
        raise ValueError(f"Unknown mode: {mode}")
    session.fs.mkdir(destition_location)
    ts.run()

def write_to_deltalake(ts:TaskSet, path:str, max_rows_per_file:int=1000000, max_partitions:int=4, engine:str='pyarrow', mode="append") -> TaskSet:
    ts.add_transform(
        TransformTask(
            name="write-delta-lake",
            fun=task_definitions.definition_write_delta_lake,
            args=[],
            kwargs={
                "path": path, 
                "max_rows_per_file": max_rows_per_file, 
                "max_partitions": max_partitions, 
                "engine": engine, 
                "mode": mode
            }
        )
    )
    return ts

def compact_deltalake(path:str) -> TaskSet:
    ts = Session().create_task_set()
    ts.add_transform(
        TransformTask(
            name="compact-delta-lake",
            fun=task_definitions.compact_delta_lake,
            args=[],
            kwargs={"path": path}
        )
    )
    return ts

def vacuum_deltalake(path:str) -> TaskSet:
    ts = Session().create_task_set()
    ts.add_transform(
        TransformTask(
            name="vacuum-delta-lake",
            fun=task_definitions.vacuum_delta_lake,
            args=[],
            kwargs={"path": path}
        )
    )
    return ts
