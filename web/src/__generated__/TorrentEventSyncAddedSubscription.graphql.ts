/**
 * @generated SignedSource<<8b0867646c71ab6ae43cfa510c493eb1>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type TorrentStatus = "ALLOCATING" | "DOWNLOADING" | "DOWNLOADING_METADATA" | "SEEDING" | "STOPPED" | "STOPPING" | "VERIFYING" | "%future added value";
export type TorrentEventSyncAddedSubscription$variables = Record<PropertyKey, never>;
export type TorrentEventSyncAddedSubscription$data = {
  readonly torrentAdded: {
    readonly bytesCompleted: string;
    readonly bytesTotal: string;
    readonly downloadSpeed: number;
    readonly eta: number | null | undefined;
    readonly id: string;
    readonly name: string;
    readonly peers: {
      readonly total: number;
    };
    readonly status: TorrentStatus;
    readonly uploadSpeed: number;
  };
};
export type TorrentEventSyncAddedSubscription = {
  response: TorrentEventSyncAddedSubscription$data;
  variables: TorrentEventSyncAddedSubscription$variables;
};

const node: ConcreteRequest = (function(){
var v0 = [
  {
    "alias": null,
    "args": null,
    "concreteType": "Torrent",
    "kind": "LinkedField",
    "name": "torrentAdded",
    "plural": false,
    "selections": [
      {
        "alias": null,
        "args": null,
        "kind": "ScalarField",
        "name": "id",
        "storageKey": null
      },
      {
        "alias": null,
        "args": null,
        "kind": "ScalarField",
        "name": "name",
        "storageKey": null
      },
      {
        "alias": null,
        "args": null,
        "kind": "ScalarField",
        "name": "status",
        "storageKey": null
      },
      {
        "alias": null,
        "args": null,
        "kind": "ScalarField",
        "name": "bytesCompleted",
        "storageKey": null
      },
      {
        "alias": null,
        "args": null,
        "kind": "ScalarField",
        "name": "bytesTotal",
        "storageKey": null
      },
      {
        "alias": null,
        "args": null,
        "kind": "ScalarField",
        "name": "downloadSpeed",
        "storageKey": null
      },
      {
        "alias": null,
        "args": null,
        "kind": "ScalarField",
        "name": "uploadSpeed",
        "storageKey": null
      },
      {
        "alias": null,
        "args": null,
        "kind": "ScalarField",
        "name": "eta",
        "storageKey": null
      },
      {
        "alias": null,
        "args": null,
        "concreteType": "PeerStats",
        "kind": "LinkedField",
        "name": "peers",
        "plural": false,
        "selections": [
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "total",
            "storageKey": null
          }
        ],
        "storageKey": null
      }
    ],
    "storageKey": null
  }
];
return {
  "fragment": {
    "argumentDefinitions": [],
    "kind": "Fragment",
    "metadata": null,
    "name": "TorrentEventSyncAddedSubscription",
    "selections": (v0/*: any*/),
    "type": "Subscription",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": [],
    "kind": "Operation",
    "name": "TorrentEventSyncAddedSubscription",
    "selections": (v0/*: any*/)
  },
  "params": {
    "cacheID": "175ecc49e280a9e58a61e1927cfaca39",
    "id": null,
    "metadata": {},
    "name": "TorrentEventSyncAddedSubscription",
    "operationKind": "subscription",
    "text": "subscription TorrentEventSyncAddedSubscription {\n  torrentAdded {\n    id\n    name\n    status\n    bytesCompleted\n    bytesTotal\n    downloadSpeed\n    uploadSpeed\n    eta\n    peers {\n      total\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "20f920232d073819fec206a1324d5485";

export default node;
