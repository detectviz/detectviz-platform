# WorkflowService implementation for Python ADK Runtime
# According to spec.md: Python 端是所有 AI 業務邏輯的唯一執行域
from datetime import datetime, timezone
import asyncio
import logging

from detectviz.contracts.v1 import adk_bridge_pb2 as pb
from detectviz.contracts.v1 import adk_bridge_pb2_grpc as pb_grpc
from google.rpc import status_pb2 as grpc_status

logger = logging.getLogger(__name__)

class WorkflowService(pb_grpc.WorkflowServiceServicer):
    """
    WorkflowService handles workflow execution in Python ADK Runtime.
    According to spec.md: Sequential/Parallel/Conditional/Loop workflows using ADK framework
    """
    
    def __init__(self):
        self.active_workflows = {}  # Store running workflow instances
        logger.info("WorkflowService initialized")
    
    async def TriggerWorkflow(self, request: pb.WorkflowRequest, context) -> pb.WorkflowRun:
        """Trigger a workflow execution"""
        try:
            # Generate unique run ID
            now = datetime.now(timezone.utc)
            run_id = f"workflow_run_{int(now.timestamp() * 1000)}"
            
            logger.info(f"Triggering workflow: {request.workflow_name}, run_id: {run_id}")
            
            # Create workflow run
            workflow_run = pb.WorkflowRun(
                run_id=run_id,
                status=pb.PENDING,
                output_json="",
                queued_at=pb.Timestamp(seconds=int(now.timestamp())),
                started_at=pb.Timestamp(seconds=int(now.timestamp())),
                ended_at=pb.Timestamp(seconds=0)
            )
            
            # Store workflow run
            self.active_workflows[run_id] = {
                'request': request,
                'run': workflow_run,
                'started_at': now
            }
            
            # Start workflow execution asynchronously
            asyncio.create_task(self._execute_workflow(run_id, request))
            
            # Return immediately with PENDING status
            return workflow_run
            
        except Exception as e:
            logger.error(f"Failed to trigger workflow {request.workflow_name}: {e}")
            now = datetime.now(timezone.utc)
            return pb.WorkflowRun(
                run_id=f"error_{int(now.timestamp())}",
                status=pb.FAILED,
                output_json=f'{{"error": "{str(e)}"}}',
                queued_at=pb.Timestamp(seconds=int(now.timestamp())),
                started_at=pb.Timestamp(seconds=int(now.timestamp())),
                ended_at=pb.Timestamp(seconds=int(now.timestamp()))
            )
    
    async def GetRun(self, request: pb.RunId, context) -> pb.WorkflowRun:
        """Get workflow run status and results"""
        run_id = request.id
        
        if run_id not in self.active_workflows:
            logger.warning(f"Workflow run not found: {run_id}")
            now = datetime.now(timezone.utc)
            return pb.WorkflowRun(
                run_id=run_id,
                status=pb.FAILED,
                output_json='{"error": "Workflow run not found"}',
                queued_at=pb.Timestamp(seconds=int(now.timestamp())),
                started_at=pb.Timestamp(seconds=int(now.timestamp())),
                ended_at=pb.Timestamp(seconds=int(now.timestamp()))
            )
        
        workflow_data = self.active_workflows[run_id]
        return workflow_data['run']
    
    async def _execute_workflow(self, run_id: str, request: pb.WorkflowRequest):
        """Execute workflow asynchronously"""
        try:
            workflow_data = self.active_workflows[run_id]
            workflow_run = workflow_data['run']
            
            logger.info(f"Executing workflow {request.workflow_name} with run_id {run_id}")
            
            # Update status to RUNNING
            workflow_run.status = pb.RUNNING
            now = datetime.now(timezone.utc)
            workflow_run.started_at.seconds = int(now.timestamp())
            
            # Simulate workflow execution
            # In real implementation, this would use ADK workflow framework
            await asyncio.sleep(1)  # Simulate processing time
            
            # Mock workflow result
            result = {
                "workflow_name": request.workflow_name,
                "input_processed": request.input_json,
                "execution_time_ms": 1000,
                "steps_completed": ["init", "process", "finalize"],
                "final_result": "Workflow completed successfully"
            }
            
            # Update final status
            workflow_run.status = pb.SUCCEEDED
            workflow_run.output_json = str(result)
            end_time = datetime.now(timezone.utc)
            workflow_run.ended_at.seconds = int(end_time.timestamp())
            
            logger.info(f"Workflow {request.workflow_name} completed successfully: {run_id}")
            
        except Exception as e:
            logger.error(f"Workflow execution failed for {run_id}: {e}")
            if run_id in self.active_workflows:
                workflow_run = self.active_workflows[run_id]['run']
                workflow_run.status = pb.FAILED
                workflow_run.output_json = f'{{"error": "{str(e)}"}}'
                now = datetime.now(timezone.utc)
                workflow_run.ended_at.seconds = int(now.timestamp())