async function synchronizeChapters() {
  const data = await fetch('/generated')
    .then((response) => response.json())
    .catch((error) => console.error(error));

  const container = document.getElementById('generatedPdfs');

  if (!data) {
    alert("There's no history of generated PDFs, please generate.")
  }

  container.innerHTML = '';
  const parsed = data
    .map((fileName) => parseInt(fileName.replace('.pdf', '')))
    .sort((a, b) => a - b);

  parsed.forEach(pdf => {
    const card = document.createElement('div');
    card.className = 'card';
    card.innerHTML = `
      <a href="/pdf/${pdf}.pdf">
        <h2>Chapter ${pdf}</h2>
        <p>View PDF</p>
      </a>
    `;
    container.appendChild(card);
  });
}

async function generatePdf(formData) {
  const response = await fetch('/download', {
    method: 'POST',
    body: formData
  }).catch((error) => console.error(error));

  const contentType = response.headers.get("Content-Type");

  if (contentType.includes('text/html')) {
    window.location.replace('/static/error.html');
  }

  if (contentType === 'application/pdf') {
    window.location.replace('/pdf/' + formData.get("chapter") +'.pdf')
  }

}

document.getElementById('downloadForm').addEventListener('submit', function(event) {
  event.preventDefault();
  const formData = new FormData(this);
  generatePdf(formData);
});
