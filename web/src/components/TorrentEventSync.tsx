import { useSubscription } from 'react-relay'
import { graphql, type RecordSourceSelectorProxy } from 'relay-runtime'

// Always-mounted subscriptions that keep ROOT_QUERY.torrents membership in
// sync. Field-level updates (speeds, progress) already work via normalize of
// torrentUpdated payloads; add/remove only touch a root field and never rewrite
// the list without an updater.

const TorrentAddedSubscription = graphql`
  subscription TorrentEventSyncAddedSubscription {
    torrentAdded {
      id
      name
      status
      bytesCompleted
      bytesTotal
      downloadSpeed
      uploadSpeed
      eta
      peers {
        total
      }
    }
  }
`

const TorrentRemovedSubscription = graphql`
  subscription TorrentEventSyncRemovedSubscription {
    torrentRemoved
  }
`

function appendTorrent(store: RecordSourceSelectorProxy<unknown>) {
  const added = store.getRootField('torrentAdded')
  if (!added) return

  const root = store.getRoot()
  const existing = root.getLinkedRecords('torrents') ?? []
  const addedId = added.getDataID()
  if (existing.some((record) => record != null && record.getDataID() === addedId)) {
    return
  }
  root.setLinkedRecords([...existing, added], 'torrents')
}

function removeTorrent(store: RecordSourceSelectorProxy<unknown>, data: unknown) {
  const removedId = (data as { torrentRemoved?: string } | null | undefined)?.torrentRemoved
  if (!removedId) return

  const root = store.getRoot()
  const existing = root.getLinkedRecords('torrents') ?? []
  root.setLinkedRecords(
    existing.filter(
      (record) =>
        record != null &&
        record.getDataID() !== removedId &&
        record.getValue('id') !== removedId,
    ),
    'torrents',
  )
  store.delete(removedId)
}

/**
 * Keeps the shared Relay `torrents` root list aligned with session add/remove
 * events. Mount once near the app shell so list and header stats stay live.
 */
export default function TorrentEventSync() {
  useSubscription({
    subscription: TorrentAddedSubscription,
    variables: {},
    updater: appendTorrent,
  })

  useSubscription({
    subscription: TorrentRemovedSubscription,
    variables: {},
    updater: removeTorrent,
  })

  return null
}
