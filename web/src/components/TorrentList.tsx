import { useLazyLoadQuery, useSubscription } from 'react-relay'
import { graphql } from 'relay-runtime'
import { Link } from 'react-router-dom'
import { formatBytes, formatSpeed, formatTime } from '../utils/format'

const TorrentListQuery = graphql`
  query TorrentListQuery {
    torrents {
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

const TorrentUpdatedSubscription = graphql`
  subscription TorrentListUpdatedSubscription {
    torrentUpdated {
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

const TorrentAddedSubscription = graphql`
  subscription TorrentListAddedSubscription {
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
  subscription TorrentListRemovedSubscription {
    torrentRemoved
  }
`

export default function TorrentList() {
  const data = useLazyLoadQuery(TorrentListQuery, {}, { fetchPolicy: 'store-and-network' })

  useSubscription({ subscription: TorrentUpdatedSubscription, variables: {} })
  useSubscription({ subscription: TorrentAddedSubscription, variables: {} })
  useSubscription({ subscription: TorrentRemovedSubscription, variables: {} })

  if (!data.torrents || data.torrents.length === 0) {
    return (
      <div className="hero min-h-[50vh] bg-base-100 rounded-box shadow-xl">
        <div className="hero-content text-center">
          <div className="max-w-md">
            <h1 className="text-3xl font-bold">No Torrents</h1>
            <p className="py-6">Your download list is empty. Start by adding a new torrent.</p>
            <Link to="/add" className="btn btn-primary">Add Torrent</Link>
          </div>
        </div>
      </div>
    )
  }

  return (
    <div className="grid grid-cols-1 gap-4">
        {data.torrents.map((torrent) => (
          <TorrentItem key={torrent.id} torrent={torrent} />
        ))}
    </div>
  )
}

function TorrentItem({ torrent }: { torrent: any }) {
  const progress = torrent.bytesTotal > 0 ? (Number(torrent.bytesCompleted) / Number(torrent.bytesTotal)) * 100 : 0

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'SEEDING': return 'badge-success';
      case 'DOWNLOADING': return 'badge-info';
      case 'STOPPED': return 'badge-neutral';
      case 'ERROR': return 'badge-error';
      default: return 'badge-ghost';
    }
  }

  return (
    <Link to={`/torrent/${torrent.id}`} className="card bg-base-100 shadow-md hover:shadow-lg transition-shadow border border-base-200">
      <div className="card-body p-4 sm:p-6">
        <div className="flex flex-col sm:flex-row justify-between gap-4">
          <div className="flex-1 min-w-0">
            <h3 className="card-title text-base sm:text-lg truncate mb-1" title={torrent.name}>
                {torrent.name}
            </h3>
            <div className="flex flex-wrap gap-2 items-center text-xs sm:text-sm text-base-content/70">
              <div className={`badge badge-sm ${getStatusColor(torrent.status)}`}>{torrent.status}</div>
              <span>{formatBytes(torrent.bytesCompleted)} / {formatBytes(torrent.bytesTotal)}</span>
              <span className="hidden sm:inline">•</span>
              <span>{formatSpeed(torrent.downloadSpeed)} ▼</span>
              <span>{formatSpeed(torrent.uploadSpeed)} ▲</span>
              <span className="hidden sm:inline">•</span>
              <span>ETA: {formatTime(torrent.eta)}</span>
            </div>
          </div>

          <div className="flex items-center gap-2 sm:hidden">
          </div>
        </div>

        <progress
            className={`progress w-full mt-4 ${torrent.status === 'SEEDING' || progress === 100 ? 'progress-success' : 'progress-primary'}`}
            value={progress}
            max="100">
        </progress>
      </div>
    </Link>
  )
}
