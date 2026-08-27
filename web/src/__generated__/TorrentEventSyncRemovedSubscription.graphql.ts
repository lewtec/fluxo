/**
 * @generated SignedSource<<d3d472124055a2d3434eb09dcb5aed86>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type TorrentEventSyncRemovedSubscription$variables = Record<PropertyKey, never>;
export type TorrentEventSyncRemovedSubscription$data = {
  readonly torrentRemoved: string;
};
export type TorrentEventSyncRemovedSubscription = {
  response: TorrentEventSyncRemovedSubscription$data;
  variables: TorrentEventSyncRemovedSubscription$variables;
};

const node: ConcreteRequest = (function(){
var v0 = [
  {
    "alias": null,
    "args": null,
    "kind": "ScalarField",
    "name": "torrentRemoved",
    "storageKey": null
  }
];
return {
  "fragment": {
    "argumentDefinitions": [],
    "kind": "Fragment",
    "metadata": null,
    "name": "TorrentEventSyncRemovedSubscription",
    "selections": (v0/*: any*/),
    "type": "Subscription",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": [],
    "kind": "Operation",
    "name": "TorrentEventSyncRemovedSubscription",
    "selections": (v0/*: any*/)
  },
  "params": {
    "cacheID": "8bc19464f98d8f3883e45337e6fd6cb8",
    "id": null,
    "metadata": {},
    "name": "TorrentEventSyncRemovedSubscription",
    "operationKind": "subscription",
    "text": "subscription TorrentEventSyncRemovedSubscription {\n  torrentRemoved\n}\n"
  }
};
})();

(node as any).hash = "df340447ef38bfea85b2bc88309c1eec";

export default node;
