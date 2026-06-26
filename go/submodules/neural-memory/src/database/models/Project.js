/**
 * Project Model
 * Rappresenta un progetto con la propria memoria
 */

import { DataTypes } from "sequelize";

export default (sequelize) => {
  return sequelize.define('Project', {
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
      }
    },
    path: {
      type: DataTypes.STRING,
      allowNull: false,
      unique: true,
      validate: {
        notEmpty: true
      }
    },
    description: {
      type: DataTypes.TEXT,
      allowNull: true,
      defaultValue: ''
    },
    metadata: {
      type: DataTypes.JSON,
      defaultValue: {}
    },
    // Statistiche cache
    stats: {
      type: DataTypes.JSON,
      defaultValue: {
        totalNodes: 0,
        lastActivity: null
      }
    }
  }, {
    tableName: 'projects',
    timestamps: true,
    createdAt: 'created_at',
    updatedAt: 'updated_at',
    
    indexes: [
      { fields: ['name'] },
      { fields: ['path'], unique: true }
    ]
  });
};
