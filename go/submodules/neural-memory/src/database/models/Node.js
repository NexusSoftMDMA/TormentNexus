/**
 * Node Model v2.0
 * Rappresenta un nodo di memoria nella rete neurale
 * 
 * CAMBIAMENTI v2.0:
 * - RIMOSSO: projectId (memoria unificata, non più vincolata a progetto)
 * - AGGIUNTO: sessionId (collegamento opzionale a sessione di lavoro)
 */

import { DataTypes } from "sequelize";

export default (sequelize) => {
  return sequelize.define('Node', {
    id: {
      type: DataTypes.UUID,
      defaultValue: DataTypes.UUIDV4,
      primaryKey: true,
      allowNull: false
    },
    sessionId: {
      type: DataTypes.UUID,
      allowNull: true,
      comment: 'Sessione di appartenenza (opzionale - per tracciamento)'
    },
    parentId: {
      type: DataTypes.UUID,
      allowNull: true,
      comment: 'Parent node per gerarchia (null = radice)'
    },
    type: {
      type: DataTypes.STRING,
      defaultValue: 'generic',
      comment: 'Tipo: task, entity, file, concept, summary, action, generic, skill, error, edge_case, operation, convention, pattern'
    },
    keywords: {
      type: DataTypes.JSON,
      defaultValue: [],
      comment: 'Array di parole chiave per ricerca'
    },
    content: {
      type: DataTypes.TEXT,
      allowNull: true,
      comment: 'Contenuto long-text del nodo'
    },
    metadata: {
      type: DataTypes.JSON,
      defaultValue: {},
      comment: 'Dati aggiuntivi: file modificati, azioni, skill framework, etc.'
    },
    depth: {
      type: DataTypes.INTEGER,
      defaultValue: 0,
      comment: 'Livello nella gerarchia (0 = radice)'
    },
    weight: {
      type: DataTypes.FLOAT,
      defaultValue: 1.0,
      comment: 'Peso per ranking (0.1 - 10.0)'
    },
    // Conteggio keyword per ottimizzazione ricerca
    keywordCount: {
      type: DataTypes.INTEGER,
      defaultValue: 0
    },
    // Timestamps
    taskDate: {
      type: DataTypes.DATEONLY,
      allowNull: true,
      comment: 'Data del task associato'
    }
  }, {
    tableName: 'nodes',
    timestamps: true,
    createdAt: 'created_at',
    updatedAt: 'updated_at',
    
    indexes: [
      { fields: ['session_id'] },
      { fields: ['parent_id'] },
      { fields: ['type'] },
      { fields: ['session_id', 'type'] },
      { fields: ['created_at'] },
      // Indice per keyword search
      { fields: ['keyword_count'] }
    ]
  });
};
