import unittest
from publication import publication_target, wait_for_release

# 2026-10-03 20:00 UTC; :00 also remains :00 in Europe/Zurich.
BOUNDARY = 1791057600


class PublicationTests(unittest.TestCase):
    def test_prepare_at_fifty_for_next_hour(self):
        self.assertEqual(publication_target(BOUNDARY - 600), BOUNDARY)

    def test_delayed_prepare_preserves_requested_hour(self):
        self.assertEqual(publication_target(BOUNDARY + 60, str(BOUNDARY)), BOUNDARY)

    def test_hour_boundary_does_not_wait_another_hour(self):
        self.assertEqual(publication_target(BOUNDARY), BOUNDARY)

    def test_repair_targets_current_hour(self):
        self.assertEqual(publication_target(BOUNDARY + 600), BOUNDARY)

    def test_requested_late_repair_keeps_unexpired_cycle(self):
        self.assertEqual(publication_target(BOUNDARY + 30 * 60, str(BOUNDARY)), BOUNDARY)

    def test_day_rollover(self):
        midnight = BOUNDARY + 4 * 3600
        self.assertEqual(publication_target(midnight - 300), midnight)

    def test_no_early_release(self):
        clock = [BOUNDARY - 90.25]
        waits = []
        def sleep(seconds):
            waits.append(seconds)
            clock[0] += seconds
        late = wait_for_release(BOUNDARY, lambda: clock[0], sleep)
        self.assertEqual(clock[0], BOUNDARY)
        self.assertEqual(late, 0)
        self.assertTrue(all(0 < seconds <= 30 for seconds in waits))

    def test_late_ready_file_releases_immediately_without_retiming(self):
        self.assertEqual(wait_for_release(BOUNDARY, lambda: BOUNDARY + 17), 17)

    def test_expired_artifact_is_not_published(self):
        with self.assertRaises(ValueError):
            wait_for_release(BOUNDARY, lambda: BOUNDARY + 3600)

    def test_invalid_requested_targets(self):
        for target in ("invalid", str(BOUNDARY + 1), str(BOUNDARY + 3600), str(BOUNDARY - 3600), "١٧٩١٠٥٧٦٠٠"):
            with self.subTest(target=target), self.assertRaises(ValueError):
                publication_target(BOUNDARY, target)


if __name__ == "__main__":
    unittest.main()
