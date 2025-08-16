from typing import Dict, Any, List, Optional

class ResponseHistoryStore:
    """
    A simple in-memory store for agent response history.

    This class provides a basic implementation of a knowledge store for the MVP.
    In a production environment, this would be replaced with a persistent
    backend like Redis, Firestore, or a vector database.
    """

    def __init__(self):
        self._store: Dict[str, Dict[str, Any]] = {}
        print("Initialized in-memory ResponseHistoryStore.")

    async def add_response(self, incident_id: str, response: Dict[str, Any]):
        """
        Adds a post-mortem analysis response to the history store.

        Args:
            incident_id: The unique identifier for the incident.
            response: The analysis result to store.
        """
        print(f"Storing response for incident: {incident_id}")
        self._store[incident_id] = response

    async def get_response(self, incident_id: str) -> Optional[Dict[str, Any]]:
        """
        Retrieves a specific post-mortem response by incident ID.

        Args:
            incident_id: The unique identifier for the incident.

        Returns:
            The stored response, or None if not found.
        """
        return self._store.get(incident_id)

    async def get_all_history(self) -> List[Dict[str, Any]]:
        """
        Retrieves all responses from the history store.

        Returns:
            A list of all stored responses.
        """
        return list(self._store.values())
