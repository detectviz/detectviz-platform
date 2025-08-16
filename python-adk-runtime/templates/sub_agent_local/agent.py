from typing import Dict, Any
from detectviz_adk.agents.base.base_agent import BaseAgent

# To use tools, import them from their canonical location:
# from detectviz_adk.tools.data.health_aggregator import HealthAggregator
# from detectviz_adk.tools.reporting.report_generator import ReportGenerator


class SubAgent(BaseAgent):
    """
    A template for a sub-agent.
    """

    def __init__(self, name: str = "sub_agent_local"):
        super().__init__(
            name=name,
            model="gemini-1.5-flash-001",
            instruction="This is a template for a sub-agent. You should replace this instruction with a detailed description of the agent's purpose, capabilities, and how it should use its tools.",
            description="A template for a sub-agent."
        )

        # To assign tools to this agent, uncomment and use the imported tool objects
        # self.health_aggregator = HealthAggregator
        # self.report_generator = ReportGenerator

    async def execute(self, request: Dict[str, Any]) -> Dict[str, Any]:
        """
        The main execution method for the agent.
        This method should be implemented to define the agent's logic.
        """
        print(f"SubAgent received request: {request}")
        # Implement the agent's decision-making logic here.
        # For example, call a tool:
        # result = await self.health_aggregator.invoke(request)
        return {"status": "success", "message": "SubAgent executed successfully.", "input": request}
