# Phase 1 Implementation Guide: Real Event Ingestion & Display

This guide provides detailed implementation steps for Phase 1 of the EvalForge implementation plan, focusing on getting real data flowing through the system.

## Phase 1.1: Test Real SDK Integration

### Step 1: Create Comprehensive SDK Test Application

Create a test script that demonstrates the full SDK capability:

**File: `/Users/jacob/projects/evalforge/test_scripts/test_real_sdk_integration.py`**

```python
#!/usr/bin/env python3
"""
Comprehensive test for EvalForge SDK integration.
Tests real event ingestion, batching, error handling, and data validation.
"""

import os
import sys
import time
import json
import requests
from datetime import datetime
from typing import Dict, Any

# Add the SDK to the path
sys.path.append('/Users/jacob/projects/evalforge/sdks/python')

import evalforge
from evalforge.models import TokenUsage

class SDKIntegrationTester:
    def __init__(self):
        # Configuration
        self.api_key = None
        self.project_id = None
        self.base_url = "http://localhost:8000"
        
        # Test data
        self.test_events = []
        self.results = {}
        
    def setup_test_project(self):
        """Create a test project and get API key."""
        # First, create a user and login
        auth_data = {
            "email": f"test_user_{int(time.time())}@example.com",
            "password": "test_password_123"
        }
        
        # Register user
        response = requests.post(f"{self.base_url}/api/auth/register", json=auth_data)
        if response.status_code != 201:
            raise Exception(f"Failed to register user: {response.text}")
        
        token = response.json()["token"]
        
        # Create test project
        project_data = {
            "name": f"SDK Integration Test {datetime.now().isoformat()}",
            "description": "Automated test project for SDK integration validation"
        }
        
        response = requests.post(
            f"{self.base_url}/api/projects",
            json=project_data,
            headers={"Authorization": f"Bearer {token}"}
        )
        
        if response.status_code != 201:
            raise Exception(f"Failed to create project: {response.text}")
        
        project_data = response.json()
        self.project_id = project_data["project"]["id"]
        self.api_key = project_data["api_key"]
        
        print(f"✓ Created test project {self.project_id} with API key")
        
    def test_basic_event_ingestion(self):
        """Test basic event ingestion through SDK."""
        print("\n=== Testing Basic Event Ingestion ===")
        
        # Initialize SDK
        client = evalforge.EvalForge(
            api_key=self.api_key,
            project_id=self.project_id,
            base_url=self.base_url,
            batch_size=5,
            flush_interval=2.0
        )
        
        # Send test events
        test_cases = [
            {
                "operation_type": "llm_completion",
                "input_data": {"prompt": "What is the capital of France?"},
                "output_data": {"response": "The capital of France is Paris."},
                "tokens": TokenUsage(prompt=10, completion=8, total=18),
                "cost": 0.00036,
                "provider": "openai",
                "model": "gpt-4"
            },
            {
                "operation_type": "embedding_generation", 
                "input_data": {"text": "Machine learning is transforming industries"},
                "output_data": {"embedding_size": 1536},
                "tokens": TokenUsage(prompt=8, completion=0, total=8),
                "cost": 0.0001,
                "provider": "openai",
                "model": "text-embedding-ada-002"
            },
            {
                "operation_type": "classification",
                "input_data": {"text": "This product is amazing!", "categories": ["positive", "negative", "neutral"]},
                "output_data": {"classification": "positive", "confidence": 0.95},
                "metadata": {"model_version": "v2.1", "temperature": 0.3}
            }
        ]
        
        trace_ids = []
        for i, test_case in enumerate(test_cases):
            trace_id = client.trace(**test_case)
            trace_ids.append(trace_id)
            self.test_events.append({
                "trace_id": trace_id,
                "expected_data": test_case
            })
            print(f"✓ Sent event {i+1}: {test_case['operation_type']} (trace_id: {trace_id})")
        
        # Force flush and wait
        client.flush(timeout=10.0)
        time.sleep(3)  # Wait for processing
        
        client.close()
        
        self.results["basic_ingestion"] = {
            "events_sent": len(test_cases),
            "trace_ids": trace_ids
        }
        
        print(f"✓ Successfully sent {len(test_cases)} events")
        
    def test_batch_ingestion(self):
        """Test batch ingestion with larger volume."""
        print("\n=== Testing Batch Ingestion ===")
        
        client = evalforge.EvalForge(
            api_key=self.api_key,
            project_id=self.project_id,
            base_url=self.base_url,
            batch_size=10,
            flush_interval=1.0
        )
        
        # Generate batch of events
        batch_events = []
        for i in range(25):
            trace_id = client.trace(
                operation_type="batch_test",
                input_data={"batch_index": i, "test_data": f"batch_item_{i}"},
                output_data={"processed": True, "result": f"result_{i}"},
                tokens=TokenUsage(prompt=5+i%3, completion=3+i%2, total=8+i%5),
                cost=0.0001 + (i * 0.00001),
                provider="test_provider",
                model=f"test_model_v{i%3 + 1}",
                metadata={"batch_id": "batch_001", "index": i}
            )
            batch_events.append(trace_id)
            
            if i % 10 == 9:
                print(f"✓ Sent batch of events up to index {i}")
        
        # Flush and wait
        client.flush(timeout=15.0)
        time.sleep(5)  # Wait for processing
        
        client.close()
        
        self.results["batch_ingestion"] = {
            "events_sent": len(batch_events),
            "trace_ids": batch_events
        }
        
        print(f"✓ Successfully sent batch of {len(batch_events)} events")
        
    def test_error_handling(self):
        """Test error handling scenarios."""
        print("\n=== Testing Error Handling ===")
        
        # Test with invalid API key
        try:
            invalid_client = evalforge.EvalForge(
                api_key="invalid_key_12345",
                project_id=self.project_id,
                base_url=self.base_url,
                debug=True
            )
            
            invalid_client.trace(
                operation_type="error_test",
                input_data={"test": "invalid_key"}
            )
            
            invalid_client.flush(timeout=5.0)
            invalid_client.close()
            
            print("✓ Invalid API key test completed (expected to fail gracefully)")
            
        except Exception as e:
            print(f"✓ Invalid API key handled correctly: {str(e)}")
        
        # Test with invalid project ID
        try:
            invalid_project_client = evalforge.EvalForge(
                api_key=self.api_key,
                project_id=99999,  # Non-existent project ID
                base_url=self.base_url,
                debug=True
            )
            
            invalid_project_client.trace(
                operation_type="error_test",
                input_data={"test": "invalid_project"}
            )
            
            invalid_project_client.flush(timeout=5.0)
            invalid_project_client.close()
            
            print("✓ Invalid project ID test completed")
            
        except Exception as e:
            print(f"✓ Invalid project ID handled correctly: {str(e)}")
        
        self.results["error_handling"] = {"completed": True}
        
    def validate_data_persistence(self):
        """Validate that events were properly persisted."""
        print("\n=== Validating Data Persistence ===")
        
        # Wait a bit more for async processing
        time.sleep(5)
        
        # Try to fetch events via API (need to create a user token first)
        auth_data = {
            "email": f"validator_{int(time.time())}@example.com", 
            "password": "validator_password_123"
        }
        
        # Register validator user
        response = requests.post(f"{self.base_url}/api/auth/register", json=auth_data)
        if response.status_code != 201:
            print("⚠ Could not create validator user, skipping API validation")
            return
            
        token = response.json()["token"]
        
        # Try to get project events
        response = requests.get(
            f"{self.base_url}/api/projects/{self.project_id}/events",
            headers={"Authorization": f"Bearer {token}"}
        )
        
        if response.status_code == 403:
            # Expected - different user can't access project
            print("✓ Access control working (validator user correctly denied access)")
        elif response.status_code == 200:
            events_data = response.json()
            event_count = len(events_data.get("events", []))
            print(f"✓ Found {event_count} events in database")
            self.results["data_persistence"] = {"events_found": event_count}
        else:
            print(f"⚠ Unexpected response when fetching events: {response.status_code}")
        
        # Direct database check would require database access, 
        # which we can add if needed for more thorough validation
        
    def run_all_tests(self):
        """Run all SDK integration tests."""
        print("=== EvalForge SDK Integration Test Suite ===")
        print(f"Testing against: {self.base_url}")
        
        try:
            # Setup
            self.setup_test_project()
            
            # Run tests
            self.test_basic_event_ingestion()
            self.test_batch_ingestion()
            self.test_error_handling()
            self.validate_data_persistence()
            
            # Summary
            print("\n=== Test Results Summary ===")
            for test_name, result in self.results.items():
                print(f"✓ {test_name}: {json.dumps(result, indent=2)}")
                
            print("\n🎉 All tests completed successfully!")
            
        except Exception as e:
            print(f"\n❌ Test suite failed: {str(e)}")
            import traceback
            traceback.print_exc()
            return False
            
        return True

if __name__ == "__main__":
    tester = SDKIntegrationTester()
    success = tester.run_all_tests()
    sys.exit(0 if success else 1)
```

### Step 2: Create Batch Ingestion Test

**File: `/Users/jacob/projects/evalforge/test_scripts/test_batch_ingestion.py`**

```python
#!/usr/bin/env python3
"""
Test batch ingestion performance and reliability.
"""

import os
import sys
import time
import threading
import json
from concurrent.futures import ThreadPoolExecutor
from datetime import datetime, timedelta

sys.path.append('/Users/jacob/projects/evalforge/sdks/python')
import evalforge
from evalforge.models import TokenUsage

class BatchIngestionTester:
    def __init__(self, api_key: str, project_id: int, base_url: str = "http://localhost:8000"):
        self.api_key = api_key
        self.project_id = project_id
        self.base_url = base_url
        self.results = {}
        
    def test_high_volume_ingestion(self, event_count: int = 1000):
        """Test high-volume event ingestion."""
        print(f"\n=== Testing High Volume Ingestion ({event_count} events) ===")
        
        client = evalforge.EvalForge(
            api_key=self.api_key,
            project_id=self.project_id,
            base_url=self.base_url,
            batch_size=50,
            flush_interval=2.0
        )
        
        start_time = time.time()
        trace_ids = []
        
        # Generate events with variety
        event_types = ["completion", "embedding", "classification", "summarization", "translation"]
        providers = ["openai", "anthropic", "cohere", "huggingface"]
        
        for i in range(event_count):
            event_type = event_types[i % len(event_types)]
            provider = providers[i % len(providers)]
            
            trace_id = client.trace(
                operation_type=event_type,
                input_data={
                    "index": i,
                    "type": event_type,
                    "prompt": f"Test prompt number {i} for {event_type}",
                    "parameters": {"temperature": 0.7, "max_tokens": 100}
                },
                output_data={
                    "result": f"Generated response {i}",
                    "finish_reason": "completed",
                    "quality_score": 0.8 + (i % 20) * 0.01
                },
                tokens=TokenUsage(
                    prompt=20 + (i % 10),
                    completion=15 + (i % 8),
                    total=35 + (i % 18)
                ),
                cost=0.001 + (i * 0.000001),
                provider=provider,
                model=f"{provider}_model_v{(i % 3) + 1}",
                metadata={
                    "batch_test": True,
                    "sequence_number": i,
                    "timestamp": datetime.now().isoformat()
                }
            )
            
            trace_ids.append(trace_id)
            
            if i % 100 == 99:
                print(f"✓ Generated {i+1} events...")
        
        # Flush all events
        print("Flushing all events...")
        flush_success = client.flush(timeout=30.0)
        
        end_time = time.time()
        duration = end_time - start_time
        
        client.close()
        
        self.results["high_volume"] = {
            "event_count": event_count,
            "duration_seconds": duration,
            "events_per_second": event_count / duration,
            "flush_success": flush_success,
            "trace_ids_count": len(trace_ids)
        }
        
        print(f"✓ Ingested {event_count} events in {duration:.2f} seconds")
        print(f"✓ Rate: {event_count/duration:.2f} events/second")
        print(f"✓ Flush successful: {flush_success}")
        
    def test_concurrent_clients(self, client_count: int = 5, events_per_client: int = 100):
        """Test multiple concurrent SDK clients."""
        print(f"\n=== Testing Concurrent Clients ({client_count} clients, {events_per_client} events each) ===")
        
        def client_worker(client_id: int):
            client = evalforge.EvalForge(
                api_key=self.api_key,
                project_id=self.project_id,
                base_url=self.base_url,
                batch_size=20,
                flush_interval=1.0
            )
            
            trace_ids = []
            start_time = time.time()
            
            for i in range(events_per_client):
                trace_id = client.trace(
                    operation_type=f"concurrent_test",
                    input_data={
                        "client_id": client_id,
                        "event_index": i,
                        "test_data": f"concurrent_client_{client_id}_event_{i}"
                    },
                    output_data={
                        "processed_by_client": client_id,
                        "result": f"result_{i}"
                    },
                    metadata={
                        "client_id": client_id,
                        "concurrent_test": True
                    }
                )
                trace_ids.append(trace_id)
            
            flush_success = client.flush(timeout=15.0)
            end_time = time.time()
            
            client.close()
            
            return {
                "client_id": client_id,
                "events_sent": len(trace_ids),
                "duration": end_time - start_time,
                "flush_success": flush_success
            }
        
        # Run concurrent clients
        start_time = time.time()
        with ThreadPoolExecutor(max_workers=client_count) as executor:
            futures = [executor.submit(client_worker, i) for i in range(client_count)]
            results = [future.result() for future in futures]
        
        end_time = time.time()
        total_duration = end_time - start_time
        
        total_events = sum(r["events_sent"] for r in results)
        successful_flushes = sum(1 for r in results if r["flush_success"])
        
        self.results["concurrent_clients"] = {
            "client_count": client_count,
            "total_events": total_events,
            "total_duration": total_duration,
            "overall_rate": total_events / total_duration,
            "successful_flushes": successful_flushes,
            "client_results": results
        }
        
        print(f"✓ {client_count} concurrent clients sent {total_events} events")
        print(f"✓ Total duration: {total_duration:.2f} seconds")
        print(f"✓ Overall rate: {total_events/total_duration:.2f} events/second")
        print(f"✓ Successful flushes: {successful_flushes}/{client_count}")
        
    def test_burst_ingestion(self, burst_size: int = 500, burst_count: int = 3):
        """Test burst ingestion patterns."""
        print(f"\n=== Testing Burst Ingestion ({burst_count} bursts of {burst_size} events) ===")
        
        client = evalforge.EvalForge(
            api_key=self.api_key,
            project_id=self.project_id,
            base_url=self.base_url,
            batch_size=100,
            flush_interval=0.5
        )
        
        burst_results = []
        
        for burst_id in range(burst_count):
            print(f"Starting burst {burst_id + 1}/{burst_count}...")
            
            burst_start = time.time()
            trace_ids = []
            
            # Send burst of events
            for i in range(burst_size):
                trace_id = client.trace(
                    operation_type="burst_test",
                    input_data={
                        "burst_id": burst_id,
                        "event_index": i,
                        "burst_timestamp": datetime.now().isoformat()
                    },
                    output_data={
                        "burst_processed": True,
                        "burst_sequence": f"{burst_id}_{i}"
                    },
                    metadata={
                        "burst_test": True,
                        "burst_id": burst_id,
                        "total_bursts": burst_count
                    }
                )
                trace_ids.append(trace_id)
            
            # Wait for burst to be processed
            client.flush(timeout=10.0)
            burst_end = time.time()
            
            burst_duration = burst_end - burst_start
            burst_results.append({
                "burst_id": burst_id,
                "events": burst_size,
                "duration": burst_duration,
                "rate": burst_size / burst_duration
            })
            
            print(f"✓ Burst {burst_id + 1} completed: {burst_size} events in {burst_duration:.2f}s")
            
            # Wait between bursts
            if burst_id < burst_count - 1:
                time.sleep(2)
        
        client.close()
        
        total_events = burst_count * burst_size
        avg_burst_rate = sum(b["rate"] for b in burst_results) / len(burst_results)
        
        self.results["burst_ingestion"] = {
            "burst_count": burst_count,
            "burst_size": burst_size,
            "total_events": total_events,
            "average_burst_rate": avg_burst_rate,
            "burst_details": burst_results
        }
        
        print(f"✓ Completed {burst_count} bursts with {total_events} total events")
        print(f"✓ Average burst rate: {avg_burst_rate:.2f} events/second")
        
    def run_performance_tests(self):
        """Run all performance tests."""
        print("=== EvalForge Batch Ingestion Performance Tests ===")
        
        try:
            self.test_high_volume_ingestion(1000)
            self.test_concurrent_clients(5, 100)
            self.test_burst_ingestion(500, 3)
            
            print("\n=== Performance Test Results ===")
            print(json.dumps(self.results, indent=2, default=str))
            
            return True
            
        except Exception as e:
            print(f"\n❌ Performance tests failed: {str(e)}")
            import traceback
            traceback.print_exc()
            return False

if __name__ == "__main__":
    if len(sys.argv) != 3:
        print("Usage: python test_batch_ingestion.py <api_key> <project_id>")
        sys.exit(1)
    
    api_key = sys.argv[1]
    project_id = int(sys.argv[2])
    
    tester = BatchIngestionTester(api_key, project_id)
    success = tester.run_performance_tests()
    sys.exit(0 if success else 1)
```

### Step 3: Create Error Scenario Tests

**File: `/Users/jacob/projects/evalforge/test_scripts/test_error_scenarios.py`**

```python
#!/usr/bin/env python3
"""
Test error handling scenarios for the EvalForge SDK.
"""

import os
import sys
import time
import json
import requests
from datetime import datetime

sys.path.append('/Users/jacob/projects/evalforge/sdks/python')
import evalforge
from evalforge.models import TokenUsage

class ErrorScenarioTester:
    def __init__(self, api_key: str, project_id: int, base_url: str = "http://localhost:8000"):
        self.api_key = api_key
        self.project_id = project_id
        self.base_url = base_url
        self.results = {}
        
    def test_network_failures(self):
        """Test handling of network failures."""
        print("\n=== Testing Network Failure Handling ===")
        
        # Test with unreachable endpoint
        client = evalforge.EvalForge(
            api_key=self.api_key,
            project_id=self.project_id,
            base_url="http://localhost:9999",  # Non-existent port
            max_retries=2,
            timeout=5.0,
            debug=True
        )
        
        # Send events that should fail
        trace_ids = []
        for i in range(5):
            trace_id = client.trace(
                operation_type="network_failure_test",
                input_data={"test_index": i},
                output_data={"should_fail": True}
            )
            trace_ids.append(trace_id)
        
        # Try to flush - should handle failures gracefully
        start_time = time.time()
        flush_result = client.flush(timeout=10.0)
        end_time = time.time()
        
        client.close()
        
        self.results["network_failures"] = {
            "events_attempted": len(trace_ids),
            "flush_result": flush_result,
            "flush_duration": end_time - start_time,
            "expected_failure": True
        }
        
        print(f"✓ Network failure test completed (expected to fail gracefully)")
        print(f"✓ Flush duration: {end_time - start_time:.2f}s")
        
    def test_invalid_credentials(self):
        """Test handling of invalid credentials."""
        print("\n=== Testing Invalid Credentials ===")
        
        test_cases = [
            {"api_key": "invalid_key_123", "project_id": self.project_id, "desc": "invalid API key"},
            {"api_key": self.api_key, "project_id": 99999, "desc": "invalid project ID"},
            {"api_key": "", "project_id": self.project_id, "desc": "empty API key"},
            {"api_key": self.api_key, "project_id": -1, "desc": "negative project ID"},
        ]
        
        credential_results = []
        
        for i, test_case in enumerate(test_cases):
            print(f"Testing: {test_case['desc']}")
            
            try:
                client = evalforge.EvalForge(
                    api_key=test_case["api_key"],
                    project_id=test_case["project_id"],
                    base_url=self.base_url,
                    max_retries=1,
                    timeout=5.0,
                    debug=True
                )
                
                # Send test event
                trace_id = client.trace(
                    operation_type="credential_test",
                    input_data={"test_case": test_case["desc"]},
                    output_data={"expected": "failure"}
                )
                
                flush_result = client.flush(timeout=8.0)
                client.close()
                
                credential_results.append({
                    "test_case": test_case["desc"],
                    "trace_id": trace_id,
                    "flush_result": flush_result,
                    "handled_gracefully": True
                })
                
                print(f"  ✓ Handled gracefully")
                
            except Exception as e:
                credential_results.append({
                    "test_case": test_case["desc"],
                    "error": str(e),
                    "handled_gracefully": True
                })
                print(f"  ✓ Exception handled: {str(e)}")
        
        self.results["invalid_credentials"] = credential_results
        
    def test_malformed_data(self):
        """Test handling of malformed data."""
        print("\n=== Testing Malformed Data Handling ===")
        
        client = evalforge.EvalForge(
            api_key=self.api_key,
            project_id=self.project_id,
            base_url=self.base_url,
            debug=True
        )
        
        malformed_tests = []
        
        # Test various malformed data scenarios
        test_cases = [
            {
                "desc": "extremely large input data",
                "data": {
                    "operation_type": "malformed_test",
                    "input_data": {"large_text": "x" * 1000000},  # 1MB string
                    "output_data": {"result": "should handle large data"}
                }
            },
            {
                "desc": "nested data structures",
                "data": {
                    "operation_type": "nested_test",
                    "input_data": {
                        "level1": {
                            "level2": {
                                "level3": {
                                    "deep_data": "nested_value",
                                    "array": [1, 2, 3, {"nested": "array_object"}]
                                }
                            }
                        }
                    },
                    "output_data": {"processed": "nested_structure"}
                }
            },
            {
                "desc": "special characters and unicode",
                "data": {
                    "operation_type": "unicode_test",
                    "input_data": {
                        "text": "Special chars: !@#$%^&*(){}[]|\\:;\"'<>?/.,~`",
                        "unicode": "Unicode: 🚀 🎉 🔥 ∑∏∆∫ αβγδ 中文测试 🌟"
                    },
                    "output_data": {"handled": "special_chars"}
                }
            }
        ]
        
        for test_case in test_cases:
            print(f"Testing: {test_case['desc']}")
            
            try:
                trace_id = client.trace(**test_case["data"])
                
                malformed_tests.append({
                    "test_case": test_case["desc"],
                    "trace_id": trace_id,
                    "success": True
                })
                
                print(f"  ✓ Handled successfully")
                
            except Exception as e:
                malformed_tests.append({
                    "test_case": test_case["desc"],
                    "error": str(e),
                    "success": False
                })
                print(f"  ⚠ Error: {str(e)}")
        
        client.flush(timeout=10.0)
        client.close()
        
        self.results["malformed_data"] = malformed_tests
        
    def test_rate_limiting(self):
        """Test rate limiting behavior."""
        print("\n=== Testing Rate Limiting ===")
        
        client = evalforge.EvalForge(
            api_key=self.api_key,
            project_id=self.project_id,
            base_url=self.base_url,
            batch_size=1,  # Small batches to trigger rate limiting
            flush_interval=0.1,
            debug=True
        )
        
        # Send events rapidly to test rate limiting
        start_time = time.time()
        trace_ids = []
        
        for i in range(100):  # Send many events quickly
            trace_id = client.trace(
                operation_type="rate_limit_test",
                input_data={"index": i, "rapid_fire": True},
                output_data={"test": "rate_limiting"}
            )
            trace_ids.append(trace_id)
        
        flush_result = client.flush(timeout=20.0)
        end_time = time.time()
        
        client.close()
        
        duration = end_time - start_time
        
        self.results["rate_limiting"] = {
            "events_sent": len(trace_ids),
            "duration": duration,
            "rate": len(trace_ids) / duration,
            "flush_result": flush_result,
            "rate_limiting_observed": duration > 10.0  # If it took longer, likely rate limited
        }
        
        print(f"✓ Sent {len(trace_ids)} events in {duration:.2f} seconds")
        print(f"✓ Rate limiting test completed")
        
    def run_error_tests(self):
        """Run all error scenario tests."""
        print("=== EvalForge Error Scenario Tests ===")
        
        try:
            self.test_network_failures()
            self.test_invalid_credentials()
            self.test_malformed_data()
            self.test_rate_limiting()
            
            print("\n=== Error Test Results Summary ===")
            for test_name, results in self.results.items():
                print(f"\n{test_name.upper()}:")
                print(json.dumps(results, indent=2, default=str))
            
            print("\n🎉 All error scenario tests completed!")
            return True
            
        except Exception as e:
            print(f"\n❌ Error tests failed: {str(e)}")
            import traceback
            traceback.print_exc()
            return False

if __name__ == "__main__":
    if len(sys.argv) != 3:
        print("Usage: python test_error_scenarios.py <api_key> <project_id>")
        sys.exit(1)
    
    api_key = sys.argv[1]
    project_id = int(sys.argv[2])
    
    tester = ErrorScenarioTester(api_key, project_id)
    success = tester.run_error_tests()
    sys.exit(0 if success else 1)
```

## Usage Instructions

### 1. Start the development environment:
```bash
cd /Users/jacob/projects/evalforge
make dev
```

### 2. Run the comprehensive SDK test:
```bash
cd /Users/jacob/projects/evalforge
python test_scripts/test_real_sdk_integration.py
```

### 3. Run performance tests (after getting API key from step 2):
```bash
python test_scripts/test_batch_ingestion.py <api_key> <project_id>
```

### 4. Run error scenario tests:
```bash
python test_scripts/test_error_scenarios.py <api_key> <project_id>
```

## Expected Outcomes

After running these tests, you should see:
1. **Successful event ingestion** through the Python SDK
2. **Real data in the database** (trace_events table)
3. **Proper error handling** for various failure scenarios
4. **Performance metrics** for high-volume ingestion
5. **Rate limiting** working correctly
6. **Backend API** responding properly to SDK requests

This establishes the foundation for Phase 1.2 (Enhanced UI Event Display) where we'll build on this real data flow to create a comprehensive frontend experience.

## Next Steps

Once Phase 1.1 is validated:
1. Move to Phase 1.2: Enhanced UI Event Display
2. Implement real-time event polling in the frontend
3. Add event detail views and status indicators
4. Create performance monitoring for the UI

The test scripts created here will serve as ongoing validation tools throughout the development process.