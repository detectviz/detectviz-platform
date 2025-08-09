import asyncio
from templates.sub_agent_local.agent import SubAgent

def test_sub_agent_can_instantiate():
    agent = SubAgent()
    assert agent is not None

def test_sub_agent_min_run_event_loop():
    async def _run():
        agent = SubAgent()
        out = await agent.run({"ping": 1}, {"run_id": "test"})
        assert out.get("ok") is True
    asyncio.run(_run())
