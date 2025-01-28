import matplotlib
matplotlib.use('Agg')  # Use a non-GUI backend

import threading
import time
import os
import requests
import json
import networkx as nx
from flask import Flask, render_template, send_from_directory, make_response, send_file
import matplotlib.pyplot as plt
import hashlib

app = Flask(__name__)
app.config['STATIC_FOLDER'] = 'static'


def hash_node(node):
    """Generate a hash for the node (consistent with Chord hashing)."""
    return int(hashlib.sha1(node.encode()).hexdigest(), 16)

def build_chord_graph():
    """
    Traverse the Chord ring starting from a known node (e.g., localhost:5000),
    dynamically discover all nodes via successor links, and sort them by hash.
    """
    G = nx.DiGraph()
    visited = set()  # Keep track of visited nodes to avoid infinite loops
    nodes = []       # Store nodes and their hashes for sorting
    start_node = "localhost:5000"  # Starting node (node 0)

    try:
        # Start traversal from the known node
        current_node = start_node
        while current_node not in visited:
            # Fetch information from the current node
            response = requests.get(f"http://{current_node}/network")
            data = response.json()

            successor = data["successor"]
            predecessor = data["predecessor"]

            # Add the current node and its connections to the graph
            G.add_node(current_node)
            G.add_node(successor)
            G.add_node(predecessor)

            G.add_edge(current_node, successor)  # Connect to successor
            G.add_edge(predecessor, current_node)  # Connect from predecessor

            # Add the node and its hash to the list for sorting
            nodes.append({"id": current_node, "hash": hash_node(current_node)})

            # Mark the current node as visited and move to the successor
            visited.add(current_node)
            current_node = successor

        # Sort nodes by their hash values
        nodes.sort(key=lambda x: x["hash"])

        # Rebuild the graph using the sorted nodes
        sorted_G = nx.DiGraph()
        for i in range(len(nodes)):
            current_node = nodes[i]["id"]
            next_node = nodes[(i + 1) % len(nodes)]["id"]  # Circular connection
            sorted_G.add_node(current_node)
            sorted_G.add_edge(current_node, next_node)

        return sorted_G

    except (requests.exceptions.ConnectionError, requests.exceptions.JSONDecodeError) as e:
        print(f"Error while traversing the ring: {e}")

    return G

def generate_chord_plot(graph, output_path="static/graph.png"):
    """
    Generate the Chord ring plot and save it as an image.
    """
    pos = nx.circular_layout(graph)

    plt.figure(figsize=(10, 10))
    nx.draw(
        graph,
        pos,
        with_labels=True,
        node_color="skyblue",
        node_size=2000,
        font_size=10,
        font_color="black",
        edge_color="gray",
        arrowsize=20,
    )
    plt.title("Chord Ring Visualization", fontsize=16)
    plt.savefig(output_path)  # Save the plot to the specified path
    plt.close()

@app.route("/")
def index():
    # Regenerate the plot dynamically each time the page is requested
    G = build_chord_graph()
    generate_chord_plot(G)  # Save the updated plot
    return render_template("index.html")

@app.route("/static/graph.png")
def serve_graph():
    response = make_response(send_file("static/graph.png"))
    return response


# Flask route to serve the HTML template
@app.route("/index.html")
def render_index():
    return render_template("index.html")

def cyclebuild_of_the_plot():
    G = build_chord_graph()
    generate_chord_plot(G)  # Save the updated plot
    threading.Timer(1, cyclebuild_of_the_plot).start()

# Run the Flask app
if __name__ == "__main__":
    # Generate the initial plot when the app starts
    print("Generating initial Chord ring plot...")
    cyclebuild_of_the_plot()

    # Start the Flask app
    app.run(host="0.0.0.0", port=8080, debug=True)

