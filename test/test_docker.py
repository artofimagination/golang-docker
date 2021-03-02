import pytest
import json
from functionalTest import httpConnection
from common import *
import ipaddress

dataColumns = ("data", "expected")
createTestData = [
    ({
      'image-name': 'test-image:latest',
      'source-dir': './workercontainer'
    },
    "test-image:latest"),

    ({
      'image-name': 'test-image-failure:latest',
      'source-dir': './docker'
    },
    "close context.tar: file already closed: Error response from daemon: Cannot locate specified Dockerfile: Dockerfile")
]

ids=['Success', 'Failure']

@pytest.mark.parametrize(dataColumns, createTestData, ids=ids)
def test_CreateImage(httpConnection, data, expected):
  try:
    r = httpConnection.POST("/create-image", data)
  except Exception as e:
    pytest.fail(f"Failed to send POST request")
    return

  if r.text != expected:
    pytest.fail(f"Test failed\nReturned: {r.text}\nExpected: {expected}")
    return

  if deleteImage(data, httpConnection, data['image-name']) is False:
    pytest.fail(f"Failed to cleanup test")
    return

createTestData = [
    ({
      'image-name': 'test-image:latest',
      'source-dir': './workercontainer'
    },
    "test-image:latest"),

    ({
      'image-name': 'test-image-failure:latest'
    },
    "Image not found")
]

ids=['Success', 'No Image']

@pytest.mark.parametrize(dataColumns, createTestData, ids=ids)
def test_GetImage(httpConnection, data, expected):
  if createImage(data, httpConnection) is False:
    return

  try:
    r = httpConnection.GET("/get-image", {"image-name": data['image-name']})
  except Exception as e:
    pytest.fail(f"Failed to send GET request")
    return

  if r.text != expected:
    pytest.fail(f"Test failed\nReturned: {r.text}\nExpected: {expected}")
    return

  if deleteImage(data, httpConnection, data['image-name']) is False:
    pytest.fail(f"Failed to cleanup test")
    return

createTestData = [
    ({
      'image-name': 'test-image:latest',
      'source-dir': './workercontainer',
    },
    "Delete completed"),

    ({
      'image-name': 'test-image-failure:latest'
    },
    "Image not found"),

    ({
      'image-name': 'test-image:latest',
      'source-dir': './workercontainer',
      'port': "8082",
      'address': "0.0.0.0",
      'network': 'golang-docker_default'
    },
    "image is being used by running container"),
]

ids=['Success', 'No Image', 'Container still running']

@pytest.mark.parametrize(dataColumns, createTestData, ids=ids)
def test_DeleteImage(httpConnection, data, expected):
  if createImage(data, httpConnection) is False:
    return

  if 'port' in data:
    ID = createContainer(data, httpConnection)
    if ID is None:
      return

    if startContainer(data, httpConnection, ID) is False:
      stopContainer(data, httpConnection, ID)
      return

  try:
    r = httpConnection.POST("/delete-image", {"image-name": data['image-name']})
  except Exception as e:
    pytest.fail(f"Failed to send POST request")
    if 'port' in data:
      stopContainer(data, httpConnection, ID)
    return

  if ('port' not in data and r.text != expected) or ('port' in data and expected not in r.text):
    pytest.fail(f"Test failed\nReturned: {r.text}\nExpected: {expected}")

  if 'port' in data:
    stopContainer(data, httpConnection, ID) 

createTestData = [
    ({
      'image-name': 'test-image:latest',
      'source-dir': './workercontainer',
      'port': '8080',
      'address': '0.0.0.0'
    },
    "Container created"),

    ({
      'image-name': 'test-image-failure:latest',
      'port': '8080',
      'address': '0.0.0.0'
    },
    "Failed to create docker container")
]

ids=['Success', 'No Image']

@pytest.mark.parametrize(dataColumns, createTestData, ids=ids)
def test_CreateContainer(httpConnection, data, expected):
  if createImage(data, httpConnection) is False:
    return

  try:
    r = httpConnection.POST("/create-container", data)
  except Exception as e:
    pytest.fail(f"Failed to send POST request")
    return

  if r.text.split(":")[0] != expected:
    pytest.fail(f"Test failed\nReturned: {r.text}\nExpected: {expected}")
    return
  
  if deleteImage(data, httpConnection, data['image-name']) is False:
    pytest.fail(f"Failed to cleanup test")
    return

createTestData = [
    ({
      'image-name': 'test-image:latest',
      'source-dir': './workercontainer',
      'port': '8080',
      'address': '0.0.0.0'
    },
    "Container found"),

    ({
      'id': '1234',
    },
    "Container not found")
]

ids=['Success', 'No container']

@pytest.mark.parametrize(dataColumns, createTestData, ids=ids)
def test_GetContainer(httpConnection, data, expected):
  if createImage(data, httpConnection) is False:
    return
  ID = createContainer(data, httpConnection) 
  if ID is None:
    return

  try:
    r = httpConnection.GET("/get-container", {"id":ID})
  except Exception as e:
    pytest.fail(f"Failed to send GET request")
    return

  if r.text != expected:
    pytest.fail(f"Test failed\nReturned: {r.text}\nExpected: {expected}")
    return
  
  if 'image-name'in data and deleteImage(data, httpConnection, data['image-name']) is False:
    pytest.fail(f"Failed to cleanup test")
    return

createTestData = [
    ({
      'image-name': 'test-image:latest',
      'source-dir': './workercontainer',
      'port': '8080',
      'address': '0.0.0.0',
      'network': 'golang-docker_default'
    },
    "Container started"),

    ({
      'id': '1234',
      'network': 'golang-docker_default'
    },
    "Failed to start container: Error response from daemon: No such container: 1234")
]

ids=['Success', 'No container']

@pytest.mark.parametrize(dataColumns, createTestData, ids=ids)
def test_StartContainer(httpConnection, data, expected):
  if createImage(data, httpConnection) is False:
    return
  ID = createContainer(data, httpConnection) 
  if ID is None:
    return

  try:
    r = httpConnection.GET("/start-container", {"id":ID, "network":data["network"]})
  except Exception as e:
    pytest.fail(f"Failed to send GET request")
    return

  if r.text != expected:
    pytest.fail(f"Test failed\nReturned: {r.text}\nExpected: {expected}")
    return

  if 'id' not in data:
    stopContainer(data, httpConnection, ID)

  if 'image-name'in data and deleteImage(data, httpConnection, data['image-name']) is False:
    pytest.fail(f"Failed to cleanup test")
    return

createTestData = [
    ({
      'image-name': 'test-image:latest',
      'source-dir': './workercontainer',
      'port': '8080',
      'address': '0.0.0.0',
      'network': 'golang-docker_default'
    },
    "Container stopped"),

    ({
      'image-name': 'test-image:latest',
      'source-dir': './workercontainer',
      'port': '8080',
      'address': '0.0.0.0',
      'skip-start':1,
      'network': 'golang-docker_default'
    },
    "Container stopped")
]

ids=['Success', 'Not running']

@pytest.mark.parametrize(dataColumns, createTestData, ids=ids)
def test_StopContainer(httpConnection, data, expected):
  if createImage(data, httpConnection) is False:
    return
  ID = createContainer(data, httpConnection) 
  if ID is None:
    return

  if 'skip-start' not in data and startContainer(data, httpConnection, ID) is False:
    stopContainer(data, httpConnection, ID)
    return

  try:
    r = httpConnection.GET("/stop-container", {"id":ID})
  except Exception as e:
    pytest.fail(f"Failed to send GET request")
    return

  if r.text != expected:
    pytest.fail(f"Test failed\nReturned: {r.text}\nExpected: {expected}")
    return

  if 'image-name'in data and deleteImage(data, httpConnection, data['image-name']) is False:
    pytest.fail(f"Failed to cleanup test")
    return

createTestData = [
    ({
      'image-name': 'test-image:latest',
      'source-dir': './workercontainer',
      'port': '8080',
      'address': '0.0.0.0',
      'network': 'golang-docker_default'
    },
    "Container stopped"),

    ({
      'image-name': 'test-image:latest',
      'source-dir': './workercontainer',
      'port': '8080',
      'address': '0.0.0.0',
      'skip-start':1,
      'network': 'golang-docker_default'
    },
    "Container stopped")
]

ids=['Success', 'Not running']

@pytest.mark.parametrize(dataColumns, createTestData, ids=ids)
def test_StopContainerByImageID(httpConnection, data, expected):
  if createImage(data, httpConnection) is False:
    return
  ID = createContainer(data, httpConnection) 
  if ID is None:
    return

  if 'skip-start' not in data and startContainer(data, httpConnection, ID) is False:
    stopContainer(data, httpConnection, ID)
    return

  try:
    r = httpConnection.GET("/get-image-id-by-tag", {"image-name":data["image-name"]})
  except Exception as e:
    pytest.fail(f"Failed to send GET request")
    return

  if r.status_code != 200:
    pytest.fail(f"Failed to execute request.\nDetails: {r.text}")
    return

  ID = r.text 

  try:
    r = httpConnection.GET("/stop-container-by-image-id", {"id":ID})
  except Exception as e:
    pytest.fail(f"Failed to send GET request")
    return

  if r.text != expected:
    pytest.fail(f"Test failed\nReturned: {r.text}\nExpected: {expected}")
    return

  if 'image-name'in data and deleteImage(data, httpConnection, data['image-name']) is False:
    pytest.fail(f"Failed to cleanup test")
    return

createTestData = [
    ({
      'image-name': 'test-image:latest',
      'source-dir': './workercontainer',
      'port': '8080',
      'address': '0.0.0.0'
    },
    "Container deleted"),

    ({
      'id': '1234',
    },
    "Error response from daemon: No such container: 1234")
]

ids=['Success', 'No container']

@pytest.mark.parametrize(dataColumns, createTestData, ids=ids)
def test_DeleteContainer(httpConnection, data, expected):
  if createImage(data, httpConnection) is False:
    return

  ID = createContainer(data, httpConnection) 
  if ID is None:
    return

  try:
    r = httpConnection.POST("/delete-container", {"id":ID})
  except Exception as e:
    pytest.fail(f"Failed to send POST request")
    return

  if r.text != expected:
    pytest.fail(f"Test failed\nReturned: {r.text}\nExpected: {expected}")
    return

  if 'image-name'in data and deleteImage(data, httpConnection, data['image-name']) is False:
    pytest.fail(f"Failed to cleanup test")
    return

createTestData = [
    ({
      'image-name': 'test-image:latest',
      'source-dir': './workercontainer',
      'port': '8080',
      'address': '0.0.0.0'
    },
    "Container exists"),

    ({
      'id': '1234',
    },
    "Container not found")
]

ids=['Exists', 'No container']

@pytest.mark.parametrize(dataColumns, createTestData, ids=ids)
def test_ContainerExists(httpConnection, data, expected):
  if createImage(data, httpConnection) is False:
    return
    
  ID = createContainer(data, httpConnection) 
  if ID is None:
    return

  try:
    r = httpConnection.POST("/container-exists", {"id":ID})
  except Exception as e:
    pytest.fail(f"Failed to send POST request")
    return

  if r.text != expected:
    pytest.fail(f"Test failed\nReturned: {r.text}\nExpected: {expected}")
    return

  if 'image-name'in data and deleteImage(data, httpConnection, data['image-name']) is False:
    pytest.fail(f"Failed to cleanup test")
    return

createTestData = [
    ({
      'image-name': 'test-image:latest',
      'source-dir': './workercontainer',
      'port': '8080',
      'address': '0.0.0.0',
      'network': 'golang-docker_default'
    },
    "172.27.0.3/16"),

    ({
      'image-name': 'test-image:latest',
      'source-dir': './workercontainer',
      'port': '8080',
      'address': '0.0.0.0',
      'network': 'golang-docker_default',
      'skip-start':1
    },
    "Container is not running or not connected to any network")
]

ids=['Exists', 'No container']

@pytest.mark.parametrize(dataColumns, createTestData, ids=ids)
def test_GetIPAddress(httpConnection, data, expected):
  if createImage(data, httpConnection) is False:
    return
    
  ID = createContainer(data, httpConnection) 
  if ID is None:
    return

  if 'skip-start' not in data and startContainer(data, httpConnection, ID) is False:
    stopContainer(data, httpConnection, ID)
    return

  try:
    r = httpConnection.GET("/get-container-ip", {"id":ID, "network":data["network"]})
  except Exception as e:
    pytest.fail(f"Failed to send POST request")
    stopContainer(data, httpConnection, ID)
    return

  try:
    ip = r.text.split('/')
    a = ipaddress.ip_address(ip[0])
  except:
    if r.text != expected:
      pytest.fail(f"Test failed\nReturned: {r.text}\nExpected: {expected}")
      stopContainer(data, httpConnection, ID)
      return

  if 'image-name'in data and deleteImage(data, httpConnection, data['image-name']) is False:
    pytest.fail(f"Failed to cleanup test")
    stopContainer(data, httpConnection, ID)
    return