from detectviz.contracts.v1 import adk_bridge_pb2 as pb
from detectviz.contracts.v1 import adk_bridge_pb2_grpc as pb_grpc

class HealthService(pb_grpc.HealthServiceServicer):
    async def Check(self, request, context):
        return pb.HealthCheckResponse(status="SERVING")
