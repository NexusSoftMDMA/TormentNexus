/**
 * Link Model v2.0
 * Rappresenta un collegamento tra nodi nella rete neurale
 * 
 * CAMBIAMENTI v2.0:
 * - Rimossi prefix dai campi (fromNodeId, toNodeId)
 * - LinkType invece di type
 */

import { DataTypes } from "sequelize";

export default (sequelize) => {
  return sequelize.define('Link', {
    id: {
      type: DataTypes.UUID,
      defaultValue: DataTypes.UUIDV4,
      primaryKey: true,
      allowNull: false
    },
    fromNodeId: {
      type: DataTypes.UUID,
      allowNull: false,
      comment: 'Nodo sorgente del collegamento'
    },
    toNodeId: {
      type: DataTypes.UUID,
      allowNull: false,
      comment: 'Nodo destinazione del collegamento'
    },
    linkType: {
      type: DataTypes.STRING,
      defaultValue: 'related',
      comment: 'Tipo: child, parent, related, reference, trigger, caused'
    },
    weight: {
      type: DataTypes.FLOAT,
      defaultValue: 1.0,
      comment: 'Peso per ranking (0.1 - 10.0)'
    },
    metadata: {
      type: DataTypes.JSON,
      defaultValue: {},
      comment: 'Dati aggiuntivi del collegamento'
    }
  }, {
    tableName: 'node_links',
    timestamps: true,
    createdAt: 'created_at',
    updatedAt: 'updated_at',
    
    indexes: [
      { fields: ['from_node_id'] },
      { fields: ['to_node_id'] },
      { fields: ['link_type'] }
    ]
  });
};
