window.addEventListener('DOMContentLoaded', () => {
    const { ipcRenderer } = window.electron;
  
    let editorInstance = null;
  
    const uploadBtn = document.getElementById('uploadBtn');
    const transformBtn = document.getElementById('transformBtn');
    const deployBtn = document.getElementById('deployBtn');
    const fileInput = document.getElementById('fileInput');
    const errorMessage = document.getElementById('errorMessage');
    const modal = document.getElementById('modal');
    const closeModal = document.getElementsByClassName('close')[0];
    const newModelBtn = document.getElementById('newModelBtn');
    const uploadFileBtn = document.getElementById('uploadFileBtn');
    const outputContainer = document.getElementById('outputContainer');
  
    uploadBtn.addEventListener('click', () => {
      modal.style.display = 'block';
    });
  
    closeModal.addEventListener('click', () => {
      modal.style.display = 'none';
    });
  
    window.addEventListener('click', (event) => {
      if (event.target == modal) {
        modal.style.display = 'none';
      }
    });
  
    newModelBtn.addEventListener('click', () => {
      modal.style.display = 'none';
      errorMessage.style.display = 'none';
      openEditor('');
    });
  
    uploadFileBtn.addEventListener('click', () => {
      modal.style.display = 'none';
      fileInput.click();
    });
  
    fileInput.addEventListener('change', (event) => {
      const file = event.target.files[0];
      if (file && (file.name.endsWith('.yaml') || file.name.endsWith('.yml'))) {
        const reader = new FileReader();
        reader.onload = (e) => {
          const code = e.target.result;
          errorMessage.style.display = 'none';
          openEditor(code);
        };
        reader.readAsText(file);
      } else {
        errorMessage.textContent = 'Please select a valid .yaml or .yml file.';
        errorMessage.style.display = 'block';
      }
    });
  
    transformBtn.addEventListener('click', () => {
      if (editorInstance) {
        const modelContent = editorInstance.getValue();
        ipcRenderer.invoke('deploy-model', modelContent).then(result => {
          displayOutput(result.join('\n'));
        }).catch(error => {
          displayError(error);
        });
      }
    });
  
    deployBtn.addEventListener('click', () => {
      if (editorInstance) {
        const modelContent = editorInstance.getValue();
        ipcRenderer.invoke('deploy-from-ui', modelContent).then(result => {
          const output = result.join('\n');
          if (output.includes('Successfully transformed and deployed model.')) {
            displaySuccess(output);
          } else {
            displayOutput(output);
          }
        }).catch(error => {
          displayError(error);
        });
      }
    });
  
    function openEditor(initialCode = '') {
      const editorContainer = document.getElementById('editorContainer');
      editorContainer.style.display = 'block';
  
      if (editorInstance) {
        editorInstance.dispose();
        editorContainer.innerHTML = '';
      }
  
      require.config({ paths: { 'vs': 'https://unpkg.com/monaco-editor/min/vs' }});
      require(['vs/editor/editor.main'], function() {
        editorInstance = monaco.editor.create(editorContainer, {
          value: initialCode,
          language: 'yaml',
          theme: 'vs',
          automaticLayout: true
        });
      });
    }
  
    function displayOutput(output) {
      outputContainer.innerHTML = ''; // Clear any existing content
  
      const sections = parseOutput(output);
  
      if (Object.keys(sections).length === 0) {
        displayError(output);
        return;
      }
  
      Object.keys(sections).forEach(key => {
        const section = sections[key];
        const container = document.createElement('div');
        const header = document.createElement('h2');
        header.textContent = section.title;
        const editorDiv = document.createElement('div');
        editorDiv.className = 'editor-container';
        container.appendChild(header);
        container.appendChild(editorDiv);
        outputContainer.appendChild(container);
  
        require.config({ paths: { 'vs': 'https://unpkg.com/monaco-editor/min/vs' }});
        require(['vs/editor/editor.main'], function() {
          monaco.editor.create(editorDiv, {
            value: section.content,
            language: 'yaml',
            theme: 'vs',
            readOnly: true,
            automaticLayout: true
          });
        });
      });
    }
  
    function parseOutput(output) {
      const sections = {
        dockerCompose: { title: 'Docker Compose Model', content: '' },
        kubernetes: { title: 'Kubernetes Model', content: '' },
        terraform: { title: 'Terraform Model', content: '' }
      };
  
      let currentSection = null;
      output.split('\n').forEach(line => {
        if (line.includes('DockerCompose Transformator:')) {
          currentSection = 'dockerCompose';
        } else if (line.includes('Kubernetes Transformator:')) {
          currentSection = 'kubernetes';
        } else if (line.includes('RabbitMQ Transformator:')) {
          currentSection = 'terraform';
        } else if (currentSection) {
          sections[currentSection].content += line + '\n';
        }
      });
  
      // Remove empty sections
      Object.keys(sections).forEach(key => {
        if (!sections[key].content.trim()) {
          delete sections[key];
        }
      });
  
      return sections;
    }
  
    function displayError(error) {
      outputContainer.innerHTML = `<div class="error-message">${error}</div>`;
    }
  
    function displaySuccess(success) {
      outputContainer.innerHTML = `<div class="success-message">${success}</div>`;
    }
  });
  