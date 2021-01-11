import requests
import time

class HTTPConnector():
  def __init__(self):
    self.URL = "http://127.0.0.1:8080"
    connected = False
    timeout = 15
    while timeout > 0:
      try:
        r = self.GET("/", "")
        if r.status_code == 200:
          connected = True
        break
      except:
        timeout -= 1
        time.sleep(1)

    if connected == False:
      raise Exception("Cannot connect to test server")


  def GET(self, address, params):
    url = self.URL + address
    return requests.get(url=url, params=params)

  def POST(self, address, json):
    url = self.URL + address
    return requests.post(url=url, json=json)
