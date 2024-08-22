# EICODA Repository

Dieses Repository gehört zur Masterarbeit mit dem Titel *"Enterprise Integration as Code: Eine domänenspezifische Sprache für Messaging-basierte Pipes-and-Filters-Deployments"* von Patrick Stopper.

Erstbetreuer: Prof. Dr. Uwe Breitenbücher

Zweitbetreuer: Prof. Dr. Christian Decker

## Validierung und Evaluation: Deploymentmodelle

- **`/EICODA/fallstudien-deployments`**:
  - `fallstudieX-0.yaml`: EICODA-Deploymentmodell der Fallstudie, wie in der Arbeit beschrieben.
  - **`/fallstudien-deployments/studieX/linesOfCode-Zielmodelle-Stufe0`**: Enthält die generierten Zielmodelle für Standardtechnologien, die zur Messung der Codezeilen verwendet wurden.
  - **`/fallstudien-deployments/laststufen/fallstudieX-X.yaml`**: Enthält die jeweiligen Variationen der Standardfallstudie je nach Laststufe.
  - `setup.yaml`: Enthält das verwendete Setup für die Hosts und Filtertypen.

## EICODA (Verzeichnis: `EICODA`)

Enthält das EICODA-Deploymentsystem und die dazugehörige CLI.

- Das Deploymentsystem lässt sich auf Windows mit der Datei `eicoda.exe` und auf Linux mit dem `eicoda`-Binary starten.
- Nach dem Wechsel in das Verzeichnis `EICODA` können folgende Kommandos über die CLI ausgeführt werden:

  - **`eicoda deploy`**  
    Startet den EICODA-Deploymentprozess.  
    **Benötigte Flags:**  
      - `--path`: Gibt den Pfad zu einem EICODA-Deploymentmodell in einer YAML-Datei an.  
    **Optionale Flags:**  
      - `--measure`: Misst die Zeit des EICODA-Overheads und die Zeit für das gesamte Deployment.  
      - `--no-tf`: Verhindert die Ausführung des Terraform-Transformators und Plugins. Nützlich, wenn kein Terraform installiert ist und Pipes direkt über Filter implementiert werden sollen (EICODA-Artefakte führen ein Assert durch, sodass sie auch ohne Terraform verwendet werden können).

    **Hinweise zum Deploymentprozess:**
      - Dateien, die über eine Criteria-Konfiguration übergeben werden, müssen sich auf derselben Ebene wie das EICODA-Deploymentmodell befinden (das über `--path` übergeben wird).
      - Beim Deployment mit Docker Compose wird der Docker Compose Transformator `localhost` in `host.docker.internal` transformieren.
      - Auf der Windows-Plattform kann es zu Problemen bei der Ausführung des Terraform-Providers von cyrilgdn für RabbitMQ kommen. (Ein Fehler trat auf, wurde aber auf unerklärliche Weise wieder behoben. Auf Linux-Ubuntu läuft es ohne Probleme.)

  - **`eicoda add`**  
    Persistiert Filter- und Hosttypen, die in einer separaten Datei gespeichert werden.  
    **Benötigte Flags:**  
      - `--path`: Gibt den Pfad zu einem EICODA-Deploymentmodell in einer YAML-Datei an.

  - **`eicoda destroy`**  
    Baut alle Ressourcen ab, die in den Dateien `kubernetesModel.yaml`, `rabbitMqModel.yaml` und `docker-compose.yaml` relativ zur EICODA-Binary enthalten sind.

## EICODA Benutzeroberfläche (Verzeichnis: `EICODA-UI`)

Enthält den Code für die EICODA-GUI.

- Das Projekt kann gestartet werden, indem man in das Verzeichnis `EICODA-UI` wechselt und den Befehl `npm start` ausführt. (Beim ersten Ausführen muss zunächst `npm install` ausgeführt werden!)

## Vorimplementierte Deployment-Artefakte (Verzeichnis: `EICODA-FilterType-Artifacts`)

Enthält die in EICODA vorimplementierten Deployment-Artefakte.

- In jedem Verzeichnis befinden sich alle Dateien, die für das Erstellen des Docker-Images benötigt werden, inklusive Dockerfile.
- Alternativ können auch die folgenden Images aus den öffentlichen DockerHub-Repositories des Autors verwendet werden:

  **General:**
  - Logger: `pstopper/eicoda-logger:latest`
  - Receiver: `pstopper/eicoda-receiver:latest`
  - Sender: `pstopper/eicoda-sender:latest`

  **Routing:**
  - Aggregator: `pstopper/eicoda-aggregator:latest`
  - FlexRouter: `pstopper/eicoda-flexrouter:latest`
  - MessageFilter: `pstopper/eicoda-messagefilter:latest`
  - Resequencer: `pstopper/eicoda-resequencer:latest`
  - Splitter: `pstopper/eicoda-splitter:latest`

  **Transformation:**
  - Translator: `pstopper/eicoda-translator:latest`
  - ContentFilter: `pstopper/eicoda-contentfilter:latest`
