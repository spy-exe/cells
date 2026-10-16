import os
import unittest
import numpy as np
from gymnasium.utils.env_checker import check_env
from pettingzoo.test import parallel_api_test
from ai.environments.cells import CellsEnv, CellsParallelEnv


class EnvironmentContract(unittest.TestCase):
    def test_gymnasium_interface_and_bounds(self):
        env = CellsEnv(opponents=0, horizon=8)
        try:
            check_env(env, skip_render_check=True)
            obs, info = env.reset(seed=100001)
            self.assertEqual(obs.shape, (80,))
            self.assertTrue(env.observation_space.contains(obs))
            self.assertEqual(info["seed"], 100001)
        finally:
            env.close()

    def test_seeded_commands_replay_and_finish(self):
        first = CellsEnv(opponents=2, horizon=4)
        second = CellsEnv(opponents=2, horizon=4)
        try:
            a, _ = first.reset(seed=100021)
            b, _ = second.reset(seed=100021)
            np.testing.assert_array_equal(a, b)
            for _ in range(4):
                action = np.array([0.5, -0.5, -1, -1], dtype=np.float32)
                x, y = first.step(action), second.step(action)
                np.testing.assert_array_equal(x[0], y[0])
                self.assertEqual(x[1:], y[1:])
            self.assertTrue(x[2] or x[3])
            self.assertEqual(len(x[4]["world_hash"]), 64)
        finally:
            first.close()
            second.close()

    def test_parallel_interface_uses_the_same_native_engine(self):
        env = CellsParallelEnv(agents=2, opponents=2, horizon=8)
        try:
            parallel_api_test(env, num_cycles=20)
        finally:
            env.close()

    def test_invalid_action_and_reference_policies(self):
        env = CellsEnv(opponents=1, horizon=3)
        try:
            env.reset(seed=100031)
            with self.assertRaises(ValueError):
                env.step(np.array([np.nan, 0, 0, 0], dtype=np.float32))
            result = env.step_reference("heuristic")
            self.assertEqual(result[0].shape, (80,))
            self.assertGreaterEqual(env.policy_ns, 0)
            with self.assertRaises(RuntimeError):
                env.step_reference("unknown")
        finally:
            env.close()


if __name__ == "__main__":
    unittest.main()
