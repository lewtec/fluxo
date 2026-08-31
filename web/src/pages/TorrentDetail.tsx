import { useLazyLoadQuery, useSubscription, useMutation } from 'react-relay';
import { graphql } from 'relay-runtime';
import { useParams, useNavigate } from 'react-router-dom';
import { formatBytes, formatSpeed, formatTime } from '../utils/format';
import { Play, Square, Trash2, ArrowLeft, Activity, File as FileIcon, Globe, HardDrive } from 'lucide-react';

const TorrentDetailQuery = graphql`
  query TorrentDetailQuery($id: ID!) {
    torrent(id: $id) {
      id
      name
      status
      bytesCompleted
      bytesTotal
      downloadSpeed
      uploadSpeed
      eta
      infoHash
      addedAt
      files {
        path
        length
        bytesCompleted
      }
      peers {
        total
        incoming
        outgoing
      }
      trackers {
        url
        status
      }
    }
  }
`;

const TorrentDetailUpdatedSubscription = graphql`
  subscription TorrentDetailUpdatedSubscription($id: ID) {
    torrentUpdated(id: $id) {
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
        incoming
        outgoing
      }
      files {
        path
        length
        bytesCompleted
      }
    }
  }
`;

const RemoveTorrentMutation = graphql`
  mutation TorrentDetailRemoveMutation($id: ID!) {
    removeTorrent(id: $id)
  }
`;

const StartTorrentMutation = graphql`
  mutation TorrentDetailStartMutation($id: ID!) {
    startTorrent(id: $id)
  }
`;

const StopTorrentMutation = graphql`
  mutation TorrentDetailStopMutation($id: ID!) {
    stopTorrent(id: $id)
  }
`;

export default function TorrentDetail() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();

  if (!id) return <div>Invalid ID</div>;

  const data = useLazyLoadQuery(TorrentDetailQuery, { id }, { fetchPolicy: 'store-and-network' });
  const torrent = data.torrent;

  useSubscription({
    subscription: TorrentDetailUpdatedSubscription,
    variables: { id },
  });

  const [commitRemove, isRemoving] = useMutation(RemoveTorrentMutation);
  const [commitStart, isStarting] = useMutation(StartTorrentMutation);
  const [commitStop, isStopping] = useMutation(StopTorrentMutation);

  const handleRemove = () => {
    if (confirm('Are you sure you want to remove this torrent?')) {
      commitRemove({
        variables: { id },
        onCompleted: () => {
          navigate('/');
        },
      });
    }
  };

  const handleToggleStatus = () => {
    if (torrent?.status === 'STOPPED') {
        commitStart({ variables: { id } });
    } else {
        commitStop({ variables: { id } });
    }
  };

  if (!torrent) {
    return (
      <div className="alert alert-error">
        <span>Torrent not found</span>
        <button className="btn btn-sm" onClick={() => navigate('/')}>Go Back</button>
      </div>
    );
  }

  const progress = torrent.bytesTotal !== '0'
    ? (Number(torrent.bytesCompleted) / Number(torrent.bytesTotal)) * 100
    : 0;

  return (
    <div className="space-y-6">
      {/* Header / Actions */}
      <div className="flex flex-col md:flex-row justify-between items-start md:items-center gap-4">
        <button onClick={() => navigate('/')} className="btn btn-ghost btn-sm gap-2 pl-0">
          <ArrowLeft size={18} /> Back
        </button>
        <div className="flex gap-2">
            <button
                className={`btn btn-sm ${torrent.status === 'STOPPED' ? 'btn-success' : 'btn-warning'}`}
                onClick={handleToggleStatus}
                disabled={isStarting || isStopping}
            >
                {torrent.status === 'STOPPED' ? <Play size={16} /> : <Square size={16} />}
                {torrent.status === 'STOPPED' ? 'Start' : 'Stop'}
            </button>
            <button
                className="btn btn-sm btn-error"
                onClick={handleRemove}
                disabled={isRemoving}
            >
                <Trash2 size={16} /> Remove
            </button>
        </div>
      </div>

      {/* Main Info Card */}
      <div className="card bg-base-100 shadow-xl">
        <div className="card-body">
            <h2 className="card-title break-all">{torrent.name}</h2>
            <div className="badge badge-lg">{torrent.status}</div>

            <div className="py-4 space-y-2">
                <div className="flex justify-between text-sm">
                    <span>{formatBytes(torrent.bytesCompleted)} / {formatBytes(torrent.bytesTotal)}</span>
                    <span>{progress.toFixed(1)}%</span>
                </div>
                <progress className="progress progress-primary w-full" value={progress} max="100"></progress>
            </div>

            <div className="grid grid-cols-2 md:grid-cols-4 gap-4 text-sm">
                <div className="stat p-2 bg-base-200 rounded-box">
                    <div className="stat-title">Download Speed</div>
                    <div className="stat-value text-lg">{formatSpeed(torrent.downloadSpeed)}</div>
                </div>
                <div className="stat p-2 bg-base-200 rounded-box">
                    <div className="stat-title">Upload Speed</div>
                    <div className="stat-value text-lg">{formatSpeed(torrent.uploadSpeed)}</div>
                </div>
                <div className="stat p-2 bg-base-200 rounded-box">
                    <div className="stat-title">Peers</div>
                    <div className="stat-value text-lg">{torrent.peers.total} ({torrent.peers.incoming} in / {torrent.peers.outgoing} out)</div>
                </div>
                <div className="stat p-2 bg-base-200 rounded-box">
                    <div className="stat-title">ETA</div>
                    <div className="stat-value text-lg">{formatTime(torrent.eta)}</div>
                </div>
            </div>

            <div className="divider">Details</div>
             <div className="grid grid-cols-1 md:grid-cols-2 gap-2 text-xs opacity-70">
                <div><span className="font-bold">Hash:</span> {torrent.infoHash}</div>
                <div><span className="font-bold">Added:</span> {new Date(torrent.addedAt).toLocaleString()}</div>
            </div>
        </div>
      </div>

      {/* Tabs for Files, Trackers, etc. (Simplified as sections for now) */}

      {/* Files Section */}
      <div className="collapse collapse-arrow bg-base-100 shadow-md">
        <input type="checkbox" />
        <div className="collapse-title text-xl font-medium flex items-center gap-2">
            <FileIcon size={20} /> Files ({torrent.files.length})
        </div>
        <div className="collapse-content">
            <div className="overflow-x-auto">
                <table className="table table-xs">
                    <thead>
                        <tr>
                            <th>Path</th>
                            <th>Size</th>
                            <th>Progress</th>
                        </tr>
                    </thead>
                    <tbody>
                        {torrent.files.map((file, idx) => {
                             const fProgress = Number(file.length) > 0 ? (Number(file.bytesCompleted) / Number(file.length)) * 100 : 0;
                             return (
                                <tr key={idx}>
                                    <td className="break-all">{file.path}</td>
                                    <td className="whitespace-nowrap">{formatBytes(file.length)}</td>
                                    <td className="w-24">
                                        <progress className="progress progress-xs w-full" value={fProgress} max="100"></progress>
                                    </td>
                                </tr>
                             );
                        })}
                    </tbody>
                </table>
            </div>
        </div>
      </div>

      {/* Trackers Section */}
       <div className="collapse collapse-arrow bg-base-100 shadow-md">
        <input type="checkbox" />
        <div className="collapse-title text-xl font-medium flex items-center gap-2">
            <Globe size={20} /> Trackers ({torrent.trackers.length})
        </div>
        <div className="collapse-content">
             <div className="overflow-x-auto">
                <table className="table table-xs">
                    <thead>
                        <tr>
                            <th>URL</th>
                            <th>Status</th>
                        </tr>
                    </thead>
                    <tbody>
                        {torrent.trackers.map((tracker, idx) => (
                             <tr key={idx}>
                                <td className="break-all">{tracker.url}</td>
                                <td><div className="badge badge-xs badge-ghost">{tracker.status}</div></td>
                            </tr>
                        ))}
                    </tbody>
                </table>
            </div>
        </div>
      </div>

    </div>
  );
}
