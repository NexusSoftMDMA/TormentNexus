/**
 * Project Helper
 * Helper per gestire project_id locale con file .neural-memory-id
 */

import fs from 'fs';
import path from 'path';

const PROJECT_ID_FILENAME = '.neural-memory-id';

/**
 * Legge il project_id dal file locale
 * @param {string} projectPath - Percorso del progetto
 * @returns {string|null} - Project ID o null se non trovato
 */
function readLocalProjectId(projectPath) {
  const idFilePath = path.join(projectPath, PROJECT_ID_FILENAME);
  try {
    if (fs.existsSync(idFilePath)) {
      const content = fs.readFileSync(idFilePath, 'utf-8').trim();
      // Supporta anche formato JSON { project_id: "xxx" }
      if (content.startsWith('{')) {
        const parsed = JSON.parse(content);
        return parsed.project_id || null;
      }
      return content || null;
    }
  } catch (error) {
    console.error(`Error reading ${PROJECT_ID_FILENAME}:`, error.message);
  }
  return null;
}

/**
 * Scrive il project_id nel file locale
 * @param {string} projectPath - Percorso del progetto
 * @param {string} projectId - Project ID da salvare
 */
function writeLocalProjectId(projectPath, projectId) {
  const idFilePath = path.join(projectPath, PROJECT_ID_FILENAME);
  try {
    // Crea la directory se non esiste
    const dir = path.dirname(idFilePath);
    if (!fs.existsSync(dir)) {
      fs.mkdirSync(dir, { recursive: true });
    }
    // Scrive in formato JSON per futura estensibilità
    const content = JSON.stringify({ project_id: projectId }, null, 2);
    fs.writeFileSync(idFilePath, content, 'utf-8');
    return true;
  } catch (error) {
    console.error(`Error writing ${PROJECT_ID_FILENAME}:`, error.message);
    return false;
  }
}

export {
  PROJECT_ID_FILENAME,
  readLocalProjectId,
  writeLocalProjectId
};
