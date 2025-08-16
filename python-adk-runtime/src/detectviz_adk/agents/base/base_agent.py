from google.adk.agents import Agent


class BaseAgent(Agent):
    """Base class for all agents in the Detectviz platform.

    This class can be used to add common functionality, configurations,
    or hooks to all agents in the future, ensuring a consistent
    architectural pattern.
    """

    def __init__(self, **kwargs):
        """Initializes the BaseAgent."""
        super().__init__(**kwargs)
