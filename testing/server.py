#!/usr/bin/python3
import sys 
import socket, http.client, json
from time import sleep

CatalogAddress = "catalog.cse.nd.edu"
CatalogPort = 9097

def test_new_game(conn: socket.socket):
    '''
    Tests the newgame function
    '''
    print("Testing Newgame Function:")
    newGame = {
        "type": "new_game",
        "options": [],
        "position": "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
        "pos_id": 0
    }
    conn.sendall(json.dumps(newGame).encode())
    print(conn.recv(1024).decode())

    print("testing alternate fen string")
    newGame["position"] = "8/5NQ1/2K1PP2/pp1p4/1bp1P3/5kq1/1p4p1/3r4 w - - 0 1"
    newGame["pos_id"] = 1
    conn.sendall(json.dumps(newGame).encode())
    print(conn.recv(1024).decode())

    print("Testing options")
    newGame["options"].append("Threads 23")
    newGame["options"].append("Clear Hash")
    newGame["options"].append("Ponder true")
    newGame["options"].append("What")
    newGame["pos_id"] = 3
    conn.sendall(json.dumps(newGame).encode())
    print(conn.recv(1024).decode())

def main():
    if len(sys.argv) != 2:
        print("Usage: ./server.py <serverName>")
        return
    
    conn = http.client.HTTPConnection(CatalogAddress, CatalogPort)
    conn.request("GET", "/query.json")
    # variable to get the newest nameserver version
    last_heard_from = 0
    for server in json.loads(conn.getresponse().read()):
        if server["type"] == sys.argv[1] and server["lastheardfrom"] > last_heard_from:
            print(server)
            last_heard_from = server["lastheardfrom"]
            ip = server["address"]
            port = int(server["port"])
    conn = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    conn.connect((ip, port))
    print(ip, port)
    
    test_new_game(conn)

    sleep(1)
    data = {"type": "exit"}
    conn.sendall(json.dumps(data).encode())
    conn.close()

if __name__ == "__main__":
    main()