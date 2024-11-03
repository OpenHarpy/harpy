import unittest
from .base_test_classes import HarpyTestCase
from harpy.session import Session
from harpy.dataset import Dataset

import pandas as pd
import pyarrow as pa

def create_random_data(n:int) -> pd.DataFrame:
    return pd.DataFrame({"a": [i for i in range(n)], "b": [i for i in range(n)]})

class DatasetTest(HarpyTestCase):
    
    def __setup_env__(self):
        N = 5000
        self.df = create_random_data(N)
        ds = (
            Dataset()
            .read
            .option("distribute_on_read", True)
            .option("parallelism", 5)
            .memory(self.df)
        )
        ds.write.parquet("/Volumes/data/parquet_test/").execute()
    
    def test_memory_data_pandas(self):
        df = pd.DataFrame({"a": [1, 2, 3], "b": [4, 5, 6]})
        ds = Dataset().read.memory(df)
        df_return = ds.to_pandas()
        self.assertTrue(df.equals(df_return))
    
    def test_memory_data_arrow(self):
        table = pa.Table.from_pandas(pd.DataFrame({"a": [1, 2, 3], "b": [4, 5, 6]}))
        ds = Dataset().read.memory(table)
        table_return = ds.to_arrow()
        self.assertTrue(table.equals(table_return))
        
    def test_sql_read(self):
        query = "SELECT * FROM read_parquet('/Volumes/data/parquet_test/*')"
        ds = Dataset().read.sql(query)
        df = ds.to_pandas()
        columns = ["a", "b"]
        self.assertTrue(all([col in df.columns for col in columns]))
        count = len(df)
        self.assertNotEqual(count, 0)
    
    def test_read_parquet(self):
        ds = Dataset().read.parquet("/Volumes/data/parquet_test/")
        df = ds.to_pandas(limit_datafragments=1)
        columns = ["a", "b"]
        self.assertTrue(all([col in df.columns for col in columns]))
        count = len(df)
        self.assertNotEqual(count, 0)
    
    def test_parquet_plan(self):
        ds = Dataset().read.parquet("/Volumes/data/parquet_test/")
        plan = ds.explain(return_plan=True)
        PLAN = "--- Dataset Plan ---\n > ParquetRead(/Volumes/data/parquet_test/)\n"
        self.assertEqual(plan, PLAN)
        plan_no_return = ds.explain()
        self.assertEqual(plan_no_return, None)
    
    def test_parquet_plan_detail(self):
        ds = Dataset().read.parquet("/Volumes/data/parquet_test/")
        plan = ds.explain(detailed=True, return_plan=True)
        PLAN = "--- Dataset Plan ---\n > ParquetRead(/Volumes/data/parquet_test/)\n\n--- Taskset Plan ---\n --> Node 0: Map (read_parquet) [total_functions=5]\n"
        self.assertEqual(plan, PLAN)
    
    