import matplotlib
matplotlib.use('Agg')  # Use a non-GUI backend

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
    Build a graph representation of the Chord ring, ensuring nodes are sorted
    to avoid crossing connections.
    """
    G = nx.DiGraph()
    node_data = []

    # Step 1: Retrieve the list of nodes and their information
    for port in range(5000, 5010):
        try:
            response = requests.get(f"http://localhost:{port}/network")
            data = response.json()
            current_node = f"localhost:{port}"
            successor = data["successor"]
            predecessor = data["predecessor"]

            # Store node info along with its hash
            node_data.append({
                "id": current_node,
                "successor": successor,
                "predecessor": predecessor,
                "hash": hash_node(current_node)
            })
        except (requests.exceptions.ConnectionError, requests.exceptions.JSONDecodeError) as e:
            print(f"Error with node on port {port}: {e}")
            continue

    # Step 2: Sort nodes based on their hash values
    node_data.sort(key=lambda x: x["hash"])

    # Step 3: Add nodes and edges to the graph
    for node in node_data:
        current_node = node["id"]
        successor = node["successor"]

        # Add nodes to the graph
        G.add_node(current_node)
        G.add_node(successor)

        # Add directed edges for successor connections
        G.add_edge(current_node, successor)

    return G


# Function to plot the Chord ring
def plot_chord_ring(graph, output_path):
    plt.figure(figsize=(10, 6))
    pos = nx.circular_layout(graph)  # Circular layout for the ring
    nx.draw(
        graph,
        pos,
        with_labels=True,
        node_color="skyblue",
        edge_color="gray",
        node_size=2000,
        font_size=10,
        font_weight="bold",
    )
    nx.draw_networkx_edge_labels(
        graph,
        pos,
        edge_labels={(u, v): v for u, v in graph.edges},
        font_color="black",
    )
    plt.savefig(output_path)
    plt.close()

# Flask route to serve the graph
@app.route("/")
def index():
    G = build_chord_graph()

    # Apply a circular layout
    pos = nx.circular_layout(G)

    # Plot the graph and save it as a static file
    static_path = "static/graph.png"
    plt.figure(figsize=(10, 10))
    nx.draw(
        G,
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
    plt.savefig(static_path)  # Save to file instead of displaying a GUI
    plt.close()

    # Serve the saved plot in the HTML
    return render_template("index.html", plot_url=f"/{static_path}")


@app.route("/static/graph.png")
def serve_graph():
    response = make_response(send_file("static/graph.png"))
    response.headers["Cache-Control"] = "no-store, no-cache, must-revalidate, max-age=0"
    response.headers["Pragma"] = "no-cache"
    response.headers["Expires"] = "0"
    return response


# Flask route to serve the HTML template
@app.route("/index.html")
def render_index():
    return render_template("index.html")

# Run the Flask app
if __name__ == "__main__":
    os.makedirs(app.config['STATIC_FOLDER'], exist_ok=True)
    app.run(debug=True, port=8080)

