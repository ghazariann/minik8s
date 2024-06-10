from flask import Flask, request, jsonify
import subprocess

app = Flask(__name__)

@app.route('/setload', methods=['POST'])
def set_load():
    # Get CPU and memory from user input, defaulting to 1 CPU and 15M if not specified
    cpu = request.args.get('cpu', default="1")  # Default to 1 CPU if not specified
    memory = request.args.get('memory', default="15M")  # Default to 15 MB if not specified
    duration = request.args.get('duration', default="60")  # Default to 60 seconds for the duration

    try:
        # Execute the stress-ng command to simulate CPU and memory load
        subprocess.Popen(['stress-ng', '--vm', '1', '--vm-bytes', memory, '--cpu', cpu, '--timeout', duration])
        return jsonify({"status": "Load has been initiated", "CPU": cpu, "Memory": memory, "Duration": duration})
    except Exception as e:
        return jsonify({"error": "Failed to initiate load", "details": str(e)})

@app.route('/getparams', methods=['GET'])
def get_params():
    # Retrieve and display the current settings for memory and CPU
    cpu = request.args.get('cpu', default="1")
    memory = request.args.get('memory', default="15M")
    return jsonify({"CPU": cpu, "Memory": memory})

if __name__ == '__main__':
    app.run(host='0.0.0.0', port=5000, debug=True)

#  curl 10.32.0.3:5000/getparams
#  curl -X POST 10.32.0.1:5000/setload?memory=20M&duration=60
#  curl -X POST 10.32.0.1:5000/setload?memory=15M&duration=60