import pytest
from httpConnector import HTTPConnector

@pytest.fixture
def httpConnection():
    return HTTPConnector()