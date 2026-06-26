/**
 * Session Model
 * Rappresenta una sessione di lavoro dell'AI con la memoria
 */

import { DataTypes } from "sequelize";

export default (sequelize) => {
  return sequelize.define('Session', {
    id: {
      type: DataTypes.UUID,
      defaultValue: DataTypes.UUIDV4,
      primaryKey: true,
      allowNull: false
    },
    name: {
      type: DataTypes.STRING,
      allowNull: false,
      validate: {
        notEmpty: true
      },
      comment: 'Nome sessione: es. "refactoring API Gateway"'
    },
    description: {
      type: DataTypes.TEXT,
      allowNull: true,
      defaultValue: '',
      comment: 'Descrizione dettagliata del lavoro pianificato'
    },
    startedAt: {
      type: DataTypes.DATE,
      allowNull: false,
      defaultValue: DataTypes.NOW,
      comment: 'Timestamp inizio sessione'
    },
    endedAt: {
      type: DataTypes.DATE,
      allowNull: true,
      comment: 'Timestamp fine sessione (null = attiva)'
    },
    context: {
      type: DataTypes.JSON,
      defaultValue: {},
      comment: 'Snapshot contesto iniziale della sessione'
    },
    projectPath: {
      type: DataTypes.STRING,
      allowNull: true,
      comment: 'Percorso progetto (facoltativo, solo riferimento)'
    },
    tags: {
      type: DataTypes.JSON,
      defaultValue: [],
      comment: 'Tag per categorizzare: ["backend", "api", "fastify"]'
    },
    stats: {
      type: DataTypes.JSON,
      defaultValue: {
        nodesCreated: 0,
        skillsRegistered: 0,
        skillsUsed: 0,
        durationMinutes: 0
      },
      comment: 'Statistiche sessione aggiornate in tempo reale'
    },
    isActive: {
      type: DataTypes.BOOLEAN,
      defaultValue: true,
      comment: 'True se sessione è attualmente attiva'
    }
  }, {
    tableName: 'sessions',
    timestamps: true,
    createdAt: 'created_at',
    updatedAt: 'updated_at',
    
    indexes: [
      { fields: ['is_active'] },
      { fields: ['started_at'] },
      { fields: ['project_path'] },
      { fields: ['created_at'] }
    ]
  });
};
