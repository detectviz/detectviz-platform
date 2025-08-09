import asyncio
from templates.agent_tool_wrapper.tool import AgentTool

def test_tool_can_instantiate():
    t = AgentTool()
    assert t is not None

def test_tool_min_invoke_event_loop():
    async def _run():
        t = AgentTool()
        out = await t.invoke({"ping": 1})
        assert out.get("ok") is True
    asyncio.run(_run())
