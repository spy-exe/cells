import json
from pathlib import Path
import tempfile
import unittest
from ai.artifacts import sha256, source_identity, verify_model, write_json, promotion_decision
from ai.evaluation.compare import verify_replay
from ai.environments.cells import CellsEnv


class ArtifactIntegrity(unittest.TestCase):
    def test_tampered_model_and_dirty_source_are_rejected(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            model = root / "model.zip"
            model.write_bytes(b"trusted-test-model")
            write_json(root / "manifest.json", {"model_sha256": sha256(model), "observation": 80,
                                                "action": 4, "protocol": 1, "algorithm": "PPO"})
            self.assertEqual(verify_model(model)["protocol"], 1)
            model.write_bytes(b"changed")
            with self.assertRaises(ValueError):
                verify_model(model)
            write_json(root / "source.json", {"revision": "a"*40, "dirty": True})
            with self.assertRaises(ValueError):
                source_identity(root)

    def test_replay_reproduces_native_hash_and_detects_corruption(self):
        env = CellsEnv(opponents=2, horizon=4)
        actions = []
        try:
            env.reset(seed=100001)
            for _ in range(4):
                result = env.step_reference("heuristic")
                actions.append(env.last_actions["agent_0"])
            replay = {"seed": 100001, "opponents": 2, "horizon": 4,
                      "actions": actions, "world_hash": result[4]["world_hash"]}
            self.assertEqual(verify_replay(replay), replay["world_hash"])
            replay["world_hash"] = "0"*64
            with self.assertRaises(ValueError):
                verify_replay(replay)
        finally:
            env.close()

    def test_promotion_requires_paired_evidence_and_never_promotes_production(self):
        self.assertFalse(promotion_decision([], [], [])["lab_candidate"])
        reference = [{"seed": seed, "mass_area": 100, "survival_seconds": 60, "kills": 1,
                      "win": True, "inference_p95_us": 100} for seed in range(100001, 100101)]
        weak = [{**r, "mass_area": 50} for r in reference]
        strong = [{**r, "mass_area": 130} for r in reference]
        self.assertFalse(promotion_decision(reference, reference, weak)["lab_candidate"])
        decision = promotion_decision(reference, reference, strong)
        self.assertTrue(decision["lab_candidate"])
        self.assertFalse(decision["production_promoted"])
        strong[0]["seed"] = 1
        with self.assertRaises(ValueError):
            promotion_decision(reference, reference, strong)
