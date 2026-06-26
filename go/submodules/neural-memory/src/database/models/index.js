/**
 * Models Index v2.0
 * Esporta tutti i modelli - DB UNIFICATO, NO Project separato
 * 
 * CAMBIAMENTI v2.0:
 * - RIMOSSO: Project model (non più vincolato a progetto)
 * - AGGIUNTO: Session model (tracciamento sessioni di lavoro)
 */

import SessionModel from './Session.js';
import NodeModel from './Node.js';
import LinkModel from './Link.js';

/**
 * Inizializza tutti i modelli con Sequelize
 * @param {Sequelize} sequelize 
 * @returns {Object} - { Session, Node, Link }
 */
function initModels(sequelize) {
  const Session = SessionModel(sequelize);
  const Node = NodeModel(sequelize);
  const Link = LinkModel(sequelize);
  
  // DB unificato - niente foreign key constraints tra modelli
  // Le relazioni sono gestite via software quando necessario
  
  return { Session, Node, Link };
}

export { initModels };
