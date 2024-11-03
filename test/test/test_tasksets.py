from harpy.processing.types import (
    MapTask, ReduceTask, TransformTask, FanoutTask, BatchMapTask
)
from harpy.tasksets import TaskSet, TaskSetResults, TaskSetDefinitionError
from .base_test_classes import HarpyTestCase



# Example test case
class RawTasksetTests(HarpyTestCase):
    def test_pi_test_basic_map_reduce(self):
        # ** RAW MAP REDUCE FUNCTION DEFINITIONS **
        def map_fun(i_start:int, i_end:int, n:int) -> float:
            values = [
                4 / (1 + ((i-0.5)/n)**2)
                for i in range(i_start, i_end+1)
            ]
            return 1/n * sum(values)

        def reduce_fun(*results: float) -> float:
            return sum(results)
        ts: TaskSet = self.session.create_task_set()
        # ** RAW MAP REDUCE FUNCTION DEFINITIONS **
        
        N, step = 1000, 250
        ts.add_maps([
            MapTask(name="map", fun=map_fun, kwargs={"i_start": i, "i_end": i+step-1, "n": N})
            for i in range(1, N+1, step)
        ])
        ts.add_reduce(
            ReduceTask(name="reduce", fun=reduce_fun, kwargs={})
        )
        taskset_result: TaskSetResults = ts.collect()
        self.assertAlmostEqual(taskset_result[0], 3.1415, 3)
    
    def test_reduce_limiting(self):
        # ** RAW MAP REDUCE FUNCTION DEFINITIONS **
        def map_fun(number: int) -> int:
            return number
        
        def reduce_fun(*results: int) -> int:
            return sum(results)
        # ** RAW MAP REDUCE FUNCTION DEFINITIONS **
        
        ts: TaskSet = self.session.create_task_set()
        ts.add_maps([
            MapTask(name="map", fun=map_fun, kwargs={"number": 1})
            for i in range(5)
        ])
        ts.add_reduce(
            ReduceTask(name="reduce", fun=reduce_fun, kwargs={})
        )
        taskset_result: TaskSetResults = ts.collect()
        self.assertEqual(taskset_result[0], 5)

        ts: TaskSet = self.session.create_task_set()
        ts.add_maps([
            MapTask(name="map", fun=map_fun, kwargs={"number": 1})
            for i in range(5)
        ])
        ts.add_reduce(
            ReduceTask(name="reduce", fun=reduce_fun, kwargs={}, limit=2),
        )
        taskset_result: TaskSetResults = ts.collect()
        self.assertEqual(taskset_result[0], 2)
    
    def test_transform_tasks(self):
        # ** RAW MAP REDUCE TRANSFORM FUNCTION DEFINITIONS **
        def map_fun(number: int) -> int:
            return number
        
        def reduce_fun(*results: int) -> int:
            return sum(results)
        
        def transform_fun(result: int) -> int:
            return result * 2
        # ** RAW MAP REDUCE TRANSFORM FUNCTION DEFINITIONS **
        
        # Test functions with transform task in order Map -> Reduce -> Transform
        ts: TaskSet = self.session.create_task_set()
        ts.add_maps([
            MapTask(name="map", fun=map_fun, kwargs={"number": 2}),
            MapTask(name="map", fun=map_fun, kwargs={"number": 4}),
            MapTask(name="map", fun=map_fun, kwargs={"number": 6}),
        ])
        ts.add_reduce(
            ReduceTask(name="reduce", fun=reduce_fun, kwargs={})
        )
        ts.add_transform(
            TransformTask(name="transform", fun=transform_fun, kwargs={})
        )
        taskset_result: TaskSetResults = ts.collect()
        self.assertEqual(taskset_result[0], 24)
        # Test functions with transform task in order Map -> Transform -> Reduce
        ts: TaskSet = self.session.create_task_set()
        ts.add_maps([
            MapTask(name="map", fun=map_fun, kwargs={"number": 2}),
            MapTask(name="map", fun=map_fun, kwargs={"number": 4}),
            MapTask(name="map", fun=map_fun, kwargs={"number": 6}),
        ])
        ts.add_transform(
            TransformTask(name="transform", fun=transform_fun, kwargs={})
        )
        ts.add_reduce(
            ReduceTask(name="reduce", fun=reduce_fun, kwargs={})
        )
        taskset_result: TaskSetResults = ts.collect()
        self.assertEqual(taskset_result[0], 24)
    
    def test_fanout_tasks(self):
        # ** RAW MAP FANOUT REDUCE FUNCTION DEFINITIONS **
        def map_fun(number: int) -> int:
            return number
        
        def fanout_fun(number: int) -> int:
            return number * 2
        
        def reduce_fun(*results: int) -> int:
            return sum(results)
        # ** RAW MAP FANOUT REDUCE FUNCTION DEFINITIONS **
        
        ts: TaskSet = self.session.create_task_set()
        ts.add_maps([
            MapTask(name="map", fun=map_fun, kwargs={"number": 2}),
            MapTask(name="map", fun=map_fun, kwargs={"number": 4}),
            MapTask(name="map", fun=map_fun, kwargs={"number": 6}),
        ])
        ts.add_fanout(
            FanoutTask(name="fanout", fun=fanout_fun, kwargs={}, fanout_count=3)
        )
        ts.add_reduce(
            ReduceTask(name="reduce", fun=reduce_fun, kwargs={})
        )
        taskset_result: TaskSetResults = ts.collect()
        self.assertEqual(taskset_result[0], 72)
    
    def test_pi_batched_map_reduce(self):
        # ** RAW BATCH MAP REDUCE FUNCTION DEFINITIONS **
        def map_fun(i: int, n: int) -> float:
            return 4 / (1 + ((i-0.5)/n)**2)

        def reduce_fun(*results: list[float], n:int) -> float:
            sum_all_members = 0
            for sub_result in results:
                for result in sub_result:
                    sum_all_members += result
            return 1/n * sum_all_members
        # ** RAW BATCH MAP REDUCE FUNCTION DEFINITIONS **
        
        N = 500
        def make_map_task(i, n):
            return MapTask(name="map", fun=map_fun, kwargs={"i": i, "n": n})

        reduce_task = ReduceTask(name="reduce", fun=reduce_fun, kwargs={"n": N})

        bmt = BatchMapTask(
            name="batch_map",
            batch_size=50,
            map_tasks=[make_map_task(i, N) for i in range(1, N+1)]
        )
        ts = self.session.create_task_set()
        ts.add_batch_maps(bmt)
        ts.add_reduce(reduce_task)
        taskset_result: TaskSetResults = ts.collect()
        self.assertAlmostEqual(taskset_result[0], 3.1415, 3)

class RawTasksetPlans(HarpyTestCase):
    def test_plans_consistency(self):
        # ** RAW BATCH MAP REDUCE TRANSFORM FUNCTION DEFINITIONS **
        def map_fun(i: int, n: int) -> float:
            return 4 / (1 + ((i-0.5)/n)**2)
        
        def reduce_fun(*results: list[float], n:int) -> float:
            sum_all_members = 0
            for sub_result in results:
                for result in sub_result:
                    sum_all_members += result
            return 1/n * sum_all_members
        
        def transform_fun(result: float) -> float:
            return result * 2
        # ** RAW BATCH MAP REDUCE TRANSFORM FUNCTION DEFINITIONS **
        
        N = 500
        def make_map_task(i, n):
            return MapTask(name="map", fun=map_fun, kwargs={"i": i, "n": n})

        reduce_task = ReduceTask(name="reduce", fun=reduce_fun, kwargs={"n": N})

        bmt = BatchMapTask(
            name="batch_map",
            batch_size=50,
            map_tasks=[make_map_task(i, N) for i in range(1, N+1)]
        )
        ts = self.session.create_task_set()
        ts.add_batch_maps(bmt)
        ts.add_reduce(reduce_task)
        ts.add_transform(
            TransformTask(name="transform", fun=transform_fun, kwargs={})
        )
        plan = ts.explain(return_plan=True)
        TEST_1_PLAN = """ --> Node 0: BatchMap (batch_map) with 10 definitions\n ----> Node 1: Reduce (reduce)\n ------> Node 2: Transform (transform)\n"""
        self.assertEqual(plan, TEST_1_PLAN)
        
class RawTasksetExceptions(HarpyTestCase):
    def test_badly_defined_maps(self):
        # ** RAW MAP DEFINITIONS **
        # Test 1: Missing type hint for map input
        def map_fun_only_typed_output(i_start, i_end, n) -> float:
            return 1
        # Test 2: missing type hint for map output
        def map_fun_only_typed_input(i_start:int, i_end:int, n:int):
            return 1
        # Test 3: partially missing type hints
        def map_fun_no_typed_input(i_start, i_end, n:int) -> float:
            return 1
        # Test 4: no type hints
        def map_fun_no_typed_output(i_start, i_end, n):
            return 1
        # ** RAW MAP DEFINITIONS **
        ts = self.session.create_task_set()
        # Test 1
        with self.assertRaises(TaskSetDefinitionError):
            ts.add_maps([
                MapTask(name="map", fun=map_fun_only_typed_output, kwargs={"i_start": 1, "i_end": 250, "n": 1000}),
            ])
        # Test 2
        with self.assertRaises(TaskSetDefinitionError):
            ts.add_maps([
                MapTask(name="map", fun=map_fun_only_typed_input, kwargs={"i_start": 1, "i_end": 250, "n": 1000}),
            ])
        # Test 3
        with self.assertRaises(TaskSetDefinitionError):
            ts.add_maps([
                MapTask(name="map", fun=map_fun_no_typed_input, kwargs={"i_start": 1, "i_end": 250, "n": 1000}),
            ])
        # Test 4
        with self.assertRaises(TaskSetDefinitionError):
            ts.add_maps([
                MapTask(name="map", fun=map_fun_no_typed_output, kwargs={"i_start": 1, "i_end": 250, "n": 1000}),
            ])
    
    def test_badly_defined_reduces(self):
        # ** RAW REDUCE DEFINITIONS **
        # We first need to define a map function in order to test reduce functions
        def map_fun(i_start:int, i_end:int, n:int) -> float:
            return 1.0
        # Test 1: Missing type hint for reduce input
        def reduce_fun_only_typed_output(*results) -> int:
            return 1
        # Test 2: missing type hint for reduce output
        def reduce_fun_only_typed_input(*results: float):
            return 1
        # Test 3: not starting with *results
        def reduce_fun_no_typed_input(results: float) -> int:
            return 1
        # Test 4: not starting with *results considering output of the map function
        def reduce_fun_no_typed_output(*results: int) -> float:
            return 1
        
        # ** RAW REDUCE DEFINITIONS **
        ts = self.session.create_task_set()
        ts.add_maps([
            MapTask(name="map", fun=map_fun, kwargs={"i_start": 1, "i_end": 250, "n": 1000}),
        ])
        # Test 1
        with self.assertRaises(TaskSetDefinitionError):
            ts.add_reduce(
                ReduceTask(name="reduce", fun=reduce_fun_only_typed_output, kwargs={}),
            )
        # Test 2
        with self.assertRaises(TaskSetDefinitionError):
            ts.add_reduce(
                ReduceTask(name="reduce", fun=reduce_fun_only_typed_input, kwargs={}),
            )
        # Test 3
        with self.assertRaises(TaskSetDefinitionError):
            ts.add_reduce(
                ReduceTask(name="reduce", fun=reduce_fun_no_typed_input, kwargs={}),
            )
        # Test 4
        with self.assertRaises(TaskSetDefinitionError):
            ts.add_reduce(
                ReduceTask(name="reduce", fun=reduce_fun_no_typed_output, kwargs={}),
            )
    
    def test_badly_defined_transforms(self):
        # ** RAW TRANSFORM DEFINITIONS **
        # We first need to define a map function in order to test transform functions
        def map_fun(i_start:int, i_end:int, n:int) -> float:
            return 1.0
        # We also need to define a reduce function in order to test transform functions
        def reduce_fun(*results: float) -> float:
            return 1.0
        # Test 1: Missing type hint for transform input
        def transform_fun_only_typed_output(result) -> int:
            return 1
        # Test 2: missing type hint for transform output
        def transform_fun_only_typed_input(result: int):
            return 1
        # Test 3: incorrectly typed input
        def transform_fun_no_typed_input(result: int) -> float:
            return 1
        # Test 4: no type hints
        def transform_fun_no_typed_output(result):
            return
        # ** RAW TRANSFORM DEFINITIONS **
        
        ts = self.session.create_task_set()
        ts.add_maps([
            MapTask(name="map", fun=map_fun, kwargs={"i_start": 1, "i_end": 250, "n": 1000}),
        ])
        ts.add_reduce(
            ReduceTask(name="reduce", fun=reduce_fun, kwargs={}),
        )
        # Test 1
        with self.assertRaises(TaskSetDefinitionError):
            ts.add_transform(
                TransformTask(name="transform", fun=transform_fun_only_typed_output, kwargs={}),
            )
        # Test 2
        with self.assertRaises(TaskSetDefinitionError):
            ts.add_transform(
                TransformTask(name="transform", fun=transform_fun_only_typed_input, kwargs={}),
            )
        # Test 3
        with self.assertRaises(TaskSetDefinitionError):
            ts.add_transform(
                TransformTask(name="transform", fun=transform_fun_no_typed_input, kwargs={}),
            )
        # Test 4
        with self.assertRaises(TaskSetDefinitionError):
            ts.add_transform(
                TransformTask(name="transform", fun=transform_fun_no_typed_output, kwargs={}),
            )
    def test_badly_defined_fanouts(self):
        # ** RAW FANOUT DEFINITIONS **
        # We first need to define a map function in order to test fanout functions
        def map_fun(number: int) -> int:
            return number
        # Test 1: Missing type hint for fanout input
        def fanout_fun_only_typed_output(number) -> int:
            return 1
        # Test 2: missing type hint for fanout output
        def fanout_fun_only_typed_input(number: float):
            return 1
        # Test 3: incorrectly typed input
        def fanout_fun_no_typed_input(number: float) -> float:
            return 1
        # Test 4: no type hints
        def fanout_fun_no_typed_output(number):
            return 1
        # ** RAW FANOUT DEFINITIONS **
        
        ts = self.session.create_task_set()
        ts.add_maps([
            MapTask(name="map", fun=map_fun, kwargs={"number": 2}),
        ])
        # Test 1
        with self.assertRaises(TaskSetDefinitionError):
            ts.add_fanout(
                FanoutTask(name="fanout", fun=fanout_fun_only_typed_output, kwargs={}, fanout_count=3),
            )
        # Test 2
        with self.assertRaises(TaskSetDefinitionError):
            ts.add_fanout(
                FanoutTask(name="fanout", fun=fanout_fun_only_typed_input, kwargs={}, fanout_count=3),
            )
        # Test 3
        with self.assertRaises(TaskSetDefinitionError):
            ts.add_fanout(
                FanoutTask(name="fanout", fun=fanout_fun_no_typed_input, kwargs={}, fanout_count=3),
            )
        # Test 4
        with self.assertRaises(TaskSetDefinitionError):
            ts.add_fanout(
                FanoutTask(name="fanout", fun=fanout_fun_no_typed_output, kwargs={}, fanout_count=3),
            )
    def test_badly_defined_batch_map(self):
        # ** RAW BATCH MAP DEFINITIONS **
        # We first need to define a map function in order to test batch map functions
        def map_fun_1(i: int, n: int) -> float:
            return 1
        def map_fun_2(i: int, n: int) -> float:
            return 1
        # ** RAW BATCH MAP DEFINITIONS **
        
        ts = self.session.create_task_set()
        # Test 1 => BatchMapTask with no map tasks
        with self.assertRaises(TaskSetDefinitionError):
            ts.add_batch_maps(
                BatchMapTask(
                    name="batch_map",
                    batch_size=50,
                    map_tasks=[]
                )
            )
        # Test 2 => BatchMapTask with no batch size
        with self.assertRaises(TaskSetDefinitionError):
            ts.add_batch_maps(
                BatchMapTask(
                    name="batch_map",
                    batch_size=0,
                    map_tasks=[MapTask(name="map", fun=map_fun_1, kwargs={"i": 1, "n": 1000})]
                )
            )
        # Test 3 => BatchMapTask with inconsistent functions
        with self.assertRaises(TaskSetDefinitionError):
            ts.add_batch_maps(
                BatchMapTask(
                    name="batch_map",
                    batch_size=50,
                    map_tasks=[
                        MapTask(name="map", fun=map_fun_1, kwargs={"i": 1, "n": 1000}),
                        MapTask(name="map", fun=map_fun_2, kwargs={"i": 1, "n": 1000})
                    ]
                )
            )
        