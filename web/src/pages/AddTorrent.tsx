import { useState } from 'react';
import { useMutation } from 'react-relay';
import { graphql } from 'relay-runtime';
import { useNavigate } from 'react-router-dom';
import { ArrowLeft, Upload, Link as LinkIcon, FileText } from 'lucide-react';
// @ts-expect-error parse-torrent has no bundled type declarations
import parseTorrent, { toMagnetURI } from 'parse-torrent';
import { Buffer } from 'buffer';

const AddTorrentMutation = graphql`
  mutation AddTorrentMutation($input: AddTorrentInput!) {
    addTorrent(input: $input) {
      id
      name
    }
  }
`;

export default function AddTorrent() {
  const navigate = useNavigate();
  const [activeTab, setActiveTab] = useState<'magnet' | 'file'>('magnet');
  const [magnetUri, setMagnetUri] = useState('');
  const [file, setFile] = useState<File | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [isProcessing, setIsProcessing] = useState(false);

  const [commit, isInFlight] = useMutation(AddTorrentMutation);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    setIsProcessing(true);

    let uriToAdd = '';

    try {
        if (activeTab === 'magnet') {
            if (!magnetUri.trim()) throw new Error('Magnet URI is required');
            uriToAdd = magnetUri.trim();
        } else {
            if (!file) throw new Error('File is required');

            // Read file
            const arrayBuffer = await file.arrayBuffer();
            const buffer = Buffer.from(arrayBuffer);

            try {
                // Parse torrent file
                // IMPORTANT: parse-torrent default export is the function, named export has toMagnetURI

                const parsed = await parseTorrent(buffer);
                uriToAdd = toMagnetURI(parsed);
            } catch (err) {
                throw new Error('Invalid .torrent file: ' + (err instanceof Error ? err.message : String(err)));
            }
        }

        commit({
            variables: {
                input: {
                    uri: uriToAdd,
                }
            },
            onCompleted: (response, errors) => {
                setIsProcessing(false);
                if (errors) {
                    setError(errors[0].message);
                } else {
                    navigate('/');
                }
            },
            onError: (err) => {
                setIsProcessing(false);
                setError(err.message);
            }
        });

    } catch (err: unknown) {
        setIsProcessing(false);
        setError(err instanceof Error ? err.message : String(err));
    }
  };

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (e.target.files && e.target.files[0]) {
      setFile(e.target.files[0]);
    }
  };

  return (
    <div className="max-w-2xl mx-auto">
      <button onClick={() => navigate('/')} className="btn btn-ghost btn-sm gap-2 pl-0 mb-4">
        <ArrowLeft size={18} /> Back
      </button>

      <div className="card bg-base-100 shadow-xl">
        <div className="card-body">
            <h2 className="card-title mb-6">Add Torrent</h2>

            {/* DaisyUI 5: tabs-box (was tabs-boxed); buttons instead of bare anchors */}
            <div role="tablist" className="tabs tabs-box mb-6">
                <button
                    type="button"
                    role="tab"
                    aria-selected={activeTab === 'magnet'}
                    className={`tab ${activeTab === 'magnet' ? 'tab-active' : ''}`}
                    onClick={() => setActiveTab('magnet')}
                >
                    <LinkIcon size={16} className="mr-2" /> Magnet Link
                </button>
                <button
                    type="button"
                    role="tab"
                    aria-selected={activeTab === 'file'}
                    className={`tab ${activeTab === 'file' ? 'tab-active' : ''}`}
                    onClick={() => setActiveTab('file')}
                >
                    <FileText size={16} className="mr-2" /> .torrent File
                </button>
            </div>

            <form onSubmit={handleSubmit}>
                {activeTab === 'magnet' ? (
                    <fieldset className="fieldset w-full">
                        <legend className="fieldset-legend">Magnet URI</legend>
                        <textarea
                            className="textarea textarea-bordered h-24 font-mono text-xs w-full"
                            placeholder="magnet:?xt=urn:btih:..."
                            value={magnetUri}
                            onChange={(e) => setMagnetUri(e.target.value)}
                        />
                    </fieldset>
                ) : (
                    <fieldset className="fieldset w-full">
                        <legend className="fieldset-legend">Upload .torrent file</legend>
                        <div className="border-2 border-dashed border-base-300 rounded-box p-8 text-center hover:bg-base-200 transition-colors cursor-pointer relative">
                            <input
                                type="file"
                                accept=".torrent"
                                className="absolute inset-0 w-full h-full opacity-0 cursor-pointer"
                                onChange={handleFileChange}
                            />
                            <Upload className="mx-auto mb-2 opacity-50" />
                            <p className="text-sm opacity-70">
                                {file ? file.name : 'Click or drag file here'}
                            </p>
                        </div>
                    </fieldset>
                )}

                {error && (
                    <div className="alert alert-error mt-4 text-sm">
                        <span>{error}</span>
                    </div>
                )}

                <div className="card-actions justify-end mt-6">
                    <button
                        type="submit"
                        className="btn btn-primary"
                        disabled={isInFlight || isProcessing}
                    >
                        {(isInFlight || isProcessing) && (
                            <span className="loading loading-spinner loading-sm" />
                        )}
                        {(isInFlight || isProcessing) ? 'Adding...' : 'Add Torrent'}
                    </button>
                </div>
            </form>
        </div>
      </div>
    </div>
  );
}
