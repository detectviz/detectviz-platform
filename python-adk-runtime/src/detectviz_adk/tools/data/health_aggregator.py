# python-adk-runtime/src/detectviz_adk/tools/data/health_aggregator.py

class HealthAggregator:
    """
    健康度聚合器 - 混合架構實現
    
    核心邏輯在 Go 平台實現（高性能查詢），Python 端負責協調和業務邏輯
    """
    
    def __init__(self):
        # 使用 RemoteTool 連接 Go 服務
        self.go_health_service = RemoteTool(
            tool_id="health_aggregator",
            tool_version="1.0.0",
            grpc_endpoint="localhost:6606"
        )
        
    async def get_service_health(self, 
                                service_name: str,
                                time_range: Dict) -> Dict:
        """
        獲取服務健康度
        
        委託給 Go 服務執行高效的 InfluxDB 查詢
        """
        response = await self.go_health_service.invoke({
            "action": "get_service_health",
            "service_name": service_name,
            "start_time": time_range.get("start"),
            "end_time": time_range.get("end")
        })
        
        # Python 端進行業務邏輯處理
        return self._process_health_data(response)
    
    def _process_health_data(self, raw_data: Dict) -> Dict:
        """Python 端的業務邏輯處理"""
        # 計算 SLO/SLI
        # 生成健康評分
        # 識別異常模式
        return processed_data