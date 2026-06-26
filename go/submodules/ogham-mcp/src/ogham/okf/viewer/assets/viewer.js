(function () {
  const elements = window.__OKF_ELEMENTS__ || [];
  const cy = cytoscape({
    container: document.getElementById("graph"),
    elements: elements,
    style: [
      {
        selector: "node",
        style: {
          "background-color": "data(color)",
          "label": "data(label)",
          "color": "#f1e8d1",
          "font-size": "10px",
          "text-valign": "bottom",
          "text-margin-y": 6,
          "text-outline-color": "#0d1f17",
          "text-outline-width": 2,
          "width": "data(size)",
          "height": "data(size)",
          "border-width": 0,
        },
      },
      {
        selector: "edge",
        style: {
          "width": 1,
          "line-color": "#1f3e30",
          "target-arrow-color": "#1f3e30",
          "target-arrow-shape": "triangle",
          "curve-style": "bezier",
          "opacity": 0.7,
        },
      },
      {
        selector: "node:selected",
        style: { "border-width": 3, "border-color": "#D4A843" },
      },
    ],
    layout: { name: "cose", animate: false, padding: 30, nodeRepulsion: 8000 },
    wheelSensitivity: 0.2,
  });

  const panel = document.getElementById("panel");
  const palette = {
    Decision: "#D4A843", Memory: "#4ADE80", Pattern: "#60A5FA",
    Gotcha: "#F87171", Reference: "#A78BFA", "Topic Summary": "#FB923C",
  };

  function escapeHtml(s) {
    return s.replace(/[&<>"']/g, (c) => ({
      "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;"
    }[c]));
  }

  function renderPanel(d) {
    const color = palette[d.type] || "#94a3b8";
    const tagsHtml = (d.tags || []).map((t) =>
      `<span class="tag">${escapeHtml(t)}</span>`
    ).join("");
    panel.innerHTML = `
      <h2>${escapeHtml(d.label)}</h2>
      <span class="type-chip" style="background:${color};color:#0d1f17">${escapeHtml(d.type)}</span>
      <div class="tags">${tagsHtml || '<span class="empty">no tags</span>'}</div>
      <div class="body">${escapeHtml(d.body || "")}</div>
    `;
  }

  cy.on("tap", "node", (evt) => renderPanel(evt.target.data()));
  cy.on("tap", (evt) => {
    if (evt.target === cy) {
      panel.innerHTML = `<p class="empty">Click a node to see its body.</p>`;
    }
  });
  panel.innerHTML = `<p class="empty">Click a node to see its body.</p>`;
})();
