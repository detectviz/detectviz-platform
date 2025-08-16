# python-adk-runtime/src/detectviz_adk/agents/post_mortem/postmortem_orchestrator_agent.py
from typing import Dict, Any, List

from detectviz_adk.agents.base.base_agent import BaseAgent
from detectviz_adk.tools.data.health_aggregator import HealthAggregator
from detectviz_adk.tools.reporting.report_generator import ReportGenerator
from detectviz_adk.memory.stores.response_history_store import ResponseHistoryStore

class PostmortemOrchestratorAgent(BaseAgent):
    """
    事後複盤協調器 Agent
    根據 SRE Services MAP Phase 3 的定義，負責故障後的分析與報告生成
    """

    def __init__(self, name: str = "postmortem_orchestrator"):
        super().__init__(
            name=name,
            model="gemini-1.5-flash-001",  # Use a specific model version
            instruction=(
                "你是事後複盤協調器，負責分析故障事件並生成完整的複盤報告。"
                "你的職責包括：\n"
                "1. 根據分析計畫，調用工具收集故障期間的相關數據（指標、日誌、事件）。\n"
                "2. 分析數據，識別根本原因和影響範圍。\n"
                "3. 調用工具生成詳細報告。\n"
                "4. 存儲分析結果到知識庫以備將來參考。\n"
                "5. 提出改進建議。"
            ),
            description="處理事後複盤請求，生成故障分析報告並存儲結果。"
        )

        # Assign canonical tool instances
        self.health_aggregator = HealthAggregator
        self.report_generator = ReportGenerator
        self.history_store = ResponseHistoryStore()

    async def execute_postmortem(self, request: Dict[str, Any]) -> Dict[str, Any]:
        """
        執行事後複盤流程
        """
        incident_id = request.get("incident_id")
        time_range = request.get("time_range", {})
        affected_services = request.get("affected_services", [])

        # Step 1: 規劃複盤內容
        analysis_plan = await self._plan_analysis(incident_id, affected_services)

        # Step 2: 收集故障數據
        # In a real scenario, we would pass the plan to the tool
        incident_data = await self.health_aggregator.invoke({
            "incident_id": incident_id,
            "time_range": time_range,
            "services": affected_services,
            "metrics": analysis_plan.get("metrics_to_analyze")
        })

        # Step 3: 分析根本原因 (LLM reasoning step)
        root_cause_analysis = await self._analyze_root_cause(incident_data)

        # Step 4: 生成報告
        report_url = await self.report_generator.invoke({
            "incident_id": incident_id,
            "incident_data": incident_data,
            "analysis": root_cause_analysis
        })

        # Step 5: 更新知識庫
        analysis_result = {
            "incident_id": incident_id,
            "report_url": report_url,
            "root_cause": root_cause_analysis.get("root_cause"),
            "recommendations": root_cause_analysis.get("recommendations", []),
            "timestamp": __import__("datetime").datetime.now().isoformat()
        }
        await self.history_store.add_response(incident_id, analysis_result)


        return analysis_result

    async def _plan_analysis(self, incident_id: str, affected_services: List[str]) -> Dict:
        """規劃分析策略"""
        # MVP: 返回基本規劃
        return {
            "metrics_to_analyze": ["error_rate", "latency", "throughput"],
            "log_patterns": ["ERROR", "WARN", "timeout"],
            "correlation_window": "30m"
        }

    async def _analyze_root_cause(self, incident_data: Dict) -> Dict:
        """分析根本原因"""
        # This would typically involve a call to the LLM with the collected data.
        # For MVP, we return a simulated analysis.
        # In a real implementation, you might use self.model.generate_content(...)
        return {
            "root_cause": "配置錯誤導致服務超時",
            "contributing_factors": ["流量突增", "資源不足"],
            "impact_summary": "影響 1000 用戶，持續 30 分鐘",
            "recommendations": [
                "增加超時配置的驗證機制",
                "實施自動擴容策略"
            ],
            "lessons_learned": [
                "需要更好的配置管理流程",
                "監控告警需要更精準"
            ]
        }
