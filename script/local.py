import os
import subprocess
import signal
import sys
import time
import argparse
import json

def kill_process_by_port(port):
    process = subprocess.Popen(["lsof", "-i", ":{0}".format(port)], stdout=subprocess.PIPE, stderr=subprocess.PIPE)
    stdout, stderr = process.communicate()
    for process in str(stdout.decode("utf-8")).split("\n")[1:]:       
        data = [x for x in process.split(" ") if x != '']
        if (len(data) <= 1):
            continue
        os.kill(int(data[1]), signal.SIGKILL)

def run(num_server):
    servers = [8080+i for i in range(num_server)]
    # Generate config
    data = {
        "server": ["localhost:%d"%(port) for port in servers]
    }
    config = "config/local-%d.json"%(num_server)
    open(config, "w").write(json.dumps(data, indent=2))
    print("Generate config to", config)

    # Start server process
    for port in servers:
        kill_process_by_port(port)
        subprocess.Popen(["./cs598fts", "server", ":%d"%port], stdout=sys.stdout, stderr=sys.stderr)

    subprocess.Popen(["./script/benchmark.sh"], stdout=sys.stdout, stderr=sys.stderr)
        
    try:
        time.sleep(100000)
    except KeyboardInterrupt:
        for port in servers:
            kill_process_by_port(port)
        sys.exit()

parser = argparse.ArgumentParser()
parser.add_argument("num_server")
args = parser.parse_args()
num_server = int(args.num_server)
print("Start server locally with number", num_server)



run(num_server)