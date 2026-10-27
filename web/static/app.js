const $ = (id) => document.getElementById(id);

const form = $("add-form");
const clear = $("clear");
const jobsContainer = $("jobs");
const notice = $("notice");

async function api(path, options = {}) {
  const response = await fetch(path, {
    ...options,
    headers: {
      ...(options.body
        ? {
            "Content-Type": "application/json",
          }
        : {}),
      ...options.headers,
    },
  });

  if (!response.ok) {
    let message = `Request failed (${response.status})`;
    try {
      const body = await response.json();
      message = body.error || message;
    } catch {
      // Keep the fallback message.
    }
    throw new Error(message);
  }

  if (response.status === 204) return null;
  return response.json();
}

function element(tag, className, text) {
  const node = document.createElement(tag);
  if (className) node.className = className;
  if (text !== undefined) node.textContent = text;
  return node;
}

function render(jobs) {
  jobsContainer.replaceChildren();

  if (!jobs || jobs.length === 0) {
    jobsContainer.append(element("p", "", "No downloads yet."));
    return;
  }

  for (const job of jobs) {
    const card = element("article", "job");
    const heading = element("h3", "", job.title || job.url || `Job ${job.id}`);

    const meta = element(
      "p",
      "meta",
      `#${job.id} · ${job.mode} · ${job.status}`,
    );

    card.append(heading, meta);

    if (job.status === "downloading") {
      const progress = element("progress", "", "");
      progress.max = 100;
      progress.value = Number(job.progress) || 0;

      const detail = element(
        "p",
        "meta",
        `${progress.value.toFixed(1)}% · ${job.speed || "—"} · ETA ${job.eta || "—"}`,
      );

      card.append(progress, detail);
    }

    if (job.output) {
      card.append(element("p", "meta", `Saved: ${job.output}`));
    }

    if (job.error) {
      card.append(element("p", "error", job.error));
    }

    if (job.status === "queued" || job.status === "downloading") {
      const cancel = element("button", "secondary", "Cancel");
      cancel.type = "button";
      cancel.addEventListener("click", async () => {
        cancel.disabled = true;
        try {
          await api(`/api/jobs/${job.id}`, {
            method: "DELETE",
          });
          await refresh();
        } catch (error) {
          notice.textContent = error.message;
          cancel.disabled = false;
        }
      });
      card.append(cancel);
    }

    jobsContainer.append(card);
  }
}

async function refresh() {
  try {
    const jobs = await api("/api/jobs");
    render(jobs);
  } catch (error) {
    notice.textContent = error.message;
  }
}

form.addEventListener("submit", async (event) => {
  event.preventDefault();
  notice.textContent = "";

  const submit = form.querySelector('button[type="submit"]');
  submit.disabled = true;

  try {
    await api("/api/jobs", {
      method: "POST",
      body: JSON.stringify({
        url: $("url").value,
        mode: $("mode").value,
      }),
    });

    $("url").value = "";
    notice.textContent = "Download queued.";
    await refresh();
  } catch (error) {
    notice.textContent = error.message;
  } finally {
    submit.disabled = false;
  }
});

clear.addEventListener("click", async () => {
  const userConfirmed = confirm(
    "Warning: This will permanently delete the entire dataset. Proceed?",
  );

  if (!userConfirmed) {
    console.log("Operation cancelled by user.");
    return;
  }

  try {
    const response = await fetch("/api/jobs", {
      method: "DELETE",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
        confirm: true,
      }),
    });

    if (response.ok) {
      alert("Dataset cleared successfully!");
    } else {
      alert(`Failed to clear dataset. Status: ${response.status}`);
    }
  } catch (error) {
    console.error("Network error:", error);
    alert("A network error occurred.");
  }
});

refresh();
setInterval(refresh, 2000);
