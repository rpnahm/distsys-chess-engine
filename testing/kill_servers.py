#!/usr/bin/python3
import sys 
import socket, http.client, json

CatalogAddress = "catalog.cse.nd.edu"
CatalogPort = 9097

def main():
    '''
    Kills all of the servers that start with the inputted name'''
    if len(sys.argv) != 2:
            print("Usage: ./server.py <serverName>")
            return
        
    conn = http.client.HTTPConnection(CatalogAddress, CatalogPort)
    conn.request("GET", "/query.json")
    # variable to get the newest nameserver version
    new_conn = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    for server in json.loads(conn.getresponse().read()):
        if "type" in server.keys():
            if server["type"].startswith(sys.argv[1]):
                print(server)
                new_conn = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
                try: 
                    new_conn.connect((server["address"], int(server["port"])))
                except Exception:
                    continue
                new_conn.sendall(json.dumps({"type": "exit"}).encode())
                print(server)
    new_conn.close()
    conn.close()

if __name__ == "__main__":
     main()