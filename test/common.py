import pytest
import json

def createImage(data, httpConnection):
  if 'source-dir' in data:
    try:
      r = httpConnection.POST("/create-image", data)
    except Exception as e:
      pytest.fail(f"Failed to send POST request")
      return False

    if r.status_code != 201:
      pytest.fail(f"Failed to execute request.\nDetails: {r.text}")
      return False
  return True

def createContainer(data, httpConnection):
  if 'port' in data:
    try:
      r = httpConnection.POST("/create-container", data)
    except Exception as e:
      pytest.fail(f"Failed to send POST request")
      return None

    if r.status_code != 201:
      pytest.fail(f"Failed to execute request.\nDetails: {r.text}")
      return None

    return r.text.split(":")[1].strip()
  else:
    return data["id"]

def stopContainer(data, httpConnection, ID):
  try:
    r = httpConnection.GET("/stop-container", {"id":ID})
  except Exception as e:
    pytest.fail(f"Failed to send GET request")
    return

  if r.status_code != 200:
    pytest.fail(f"Failed to execute request.\nDetails: {r.text}")
    return

def deleteImage(data, httpConnection, ID):
  try:
    r = httpConnection.POST("/delete-image", {"image-name": ID})
  except Exception as e:
    pytest.fail(f"Failed to send POST request")
    if 'port' in data:
      stopContainer(data, httpConnection, ID)
    return False
  return True