import React, { useState, useRef } from 'react';
import Papa from 'papaparse';

function App() {
  const [files, setFiles] = useState([]);
  const [databases, setDatabases] = useState([]);
  const [benchmarkData, setBenchmarkData] = useState({});
  const [activeTab, setActiveTab] = useState('OPS');
  const fileInputRef = useRef(null);

  // Define operations to display
  const operations = [
    '[C]reate',
    '[R]ead', 
    '[U]pdate',
    '[S]can::count_all (100)',
    '[S]can::limit_id (100)',
    '[S]can::limit_all (100)',
    '[S]can::limit_count (100)',
    '[S]can::limit_start_id (100)',
    '[S]can::limit_start_all (100)',
    '[S]can::limit_start_count (100)',
    '[D]elete'
  ];

  // All available metrics
  const metrics = {
    'OPS': {
      title: 'Operations Per Second',
      description: 'Operations per second (OPS) - Higher is better',
      isHigherBetter: true,
      format: (val) => Number(val).toLocaleString(undefined, {maximumFractionDigits: 2})
    },
    'Total time': {
      title: 'Total Time',
      description: 'Wall time - Lower is better - Time from start to finish as measured by a clock',
      isHigherBetter: false,
      format: (val) => val
    },
    'Mean': {
      title: 'Mean Response Time',
      description: 'Average response time - Lower is better',
      isHigherBetter: false,
      format: (val) => val
    },
    'Max': {
      title: 'Maximum Response Time',
      description: 'Maximum response time - Lower is better',
      isHigherBetter: false,
      format: (val) => val
    },
    '99th': {
      title: '99th Percentile',
      description: '99th percentile response time - Lower is better - Top 1% slowest operations',
      isHigherBetter: false,
      format: (val) => val
    },
    '95th': {
      title: '95th Percentile',
      description: '95th percentile response time - Lower is better - Top 5% slowest operations',
      isHigherBetter: false,
      format: (val) => val
    },
    '75th': {
      title: '75th Percentile',
      description: '75th percentile response time - Lower is better',
      isHigherBetter: false,
      format: (val) => val
    },
    '50th': {
      title: '50th Percentile (Median)',
      description: 'Median response time - Lower is better',
      isHigherBetter: false,
      format: (val) => val
    },
    '25th': {
      title: '25th Percentile',
      description: '25th percentile response time - Lower is better',
      isHigherBetter: false,
      format: (val) => val
    },
    '1st': {
      title: '1st Percentile',
      description: '1st percentile response time - Lower is better',
      isHigherBetter: false,
      format: (val) => val
    },
    'Min': {
      title: 'Minimum Response Time',
      description: 'Minimum response time - Lower is better',
      isHigherBetter: false,
      format: (val) => val
    },
    'IQR': {
      title: 'Interquartile Range',
      description: 'Interquartile range (75th - 25th) - Lower is better for consistent performance',
      isHigherBetter: false,
      format: (val) => val
    },
    'CPU': {
      title: 'CPU Usage',
      description: 'Average CPU usage - Lower is better',
      isHigherBetter: false,
      format: (val) => val
    },
    'Memory': {
      title: 'Memory Usage',
      description: 'Average memory usage - Lower is better',
      isHigherBetter: false,
      format: (val) => val
    },
    'Reads': {
      title: 'Disk Reads',
      description: 'Total disk reads - Lower is better',
      isHigherBetter: false,
      format: (val) => val
    },
    'Writes': {
      title: 'Disk Writes',
      description: 'Total disk writes - Lower is better',
      isHigherBetter: false,
      format: (val) => val
    },
    'System load': {
      title: 'System Load',
      description: 'System load average - Lower is better',
      isHigherBetter: false,
      format: (val) => val
    }
  };

  // Handle file uploads
  const handleFileUpload = (e) => {
    const newFiles = Array.from(e.target.files);
    if (newFiles.length === 0) return;
    
    const updatedFiles = [...files, ...newFiles];
    setFiles(updatedFiles);
    
    // Process the files
    newFiles.forEach(file => {
      const reader = new FileReader();
      reader.onload = (e) => {
        const csv = e.target.result;
        
        // Parse CSV
        Papa.parse(csv, {
          header: true,
          skipEmptyLines: true,
          complete: (results) => {
            // Extract database name from filename
            const dbName = file.name.replace('.csv', '').replace('benchmark-', '').replace('results-', '');
            
            // Update databases list if not already included
            setDatabases(prevDBs => {
              if (!prevDBs.includes(dbName)) {
                return [...prevDBs, dbName];
              }
              return prevDBs;
            });
            
            // Process data
            const processedData = {};
            results.data.forEach(row => {
              if (row.Test) {
                // Clean up the operation name to match our display format
                let opName = row.Test;
                if (!opName.startsWith('[')) {
                  // Convert operation names like "Create" to "[C]reate"
                  if (opName.toLowerCase() === 'create') opName = '[C]reate';
                  else if (opName.toLowerCase() === 'read') opName = '[R]ead';
                  else if (opName.toLowerCase() === 'update') opName = '[U]pdate';
                  else if (opName.toLowerCase() === 'delete') opName = '[D]elete';
                  else if (opName.toLowerCase().includes('scan')) {
                    opName = `[S]${opName.substring(1)}`;
                  }
                }
                
                processedData[opName] = row;
              }
            });
            
            // Update benchmark data
            setBenchmarkData(prevData => ({
              ...prevData,
              [dbName]: processedData
            }));
          }
        });
      };
      reader.readAsText(file);
    });
    
    // Reset file input
    if (fileInputRef.current) {
      fileInputRef.current.value = '';
    }
  };

  // Parse value to numeric for comparison
  const parseValueForComparison = (value) => {
    if (value === undefined || value === null || value === '') {
      return null;
    }
    
    // If already a number, return it
    if (typeof value === 'number') {
      return value;
    }
    
    // Handle time format like "1h 7m", "5m 5s", "16s 555ms", etc.
    if (typeof value === 'string') {
      let totalMilliseconds = 0;
      
      // Hours
      const hourMatch = value.match(/(\d+)h/);
      if (hourMatch) {
        totalMilliseconds += parseInt(hourMatch[1]) * 60 * 60 * 1000;
      }
      
      // Minutes
      const minuteMatch = value.match(/(\d+)m(?!s)/);
      if (minuteMatch) {
        totalMilliseconds += parseInt(minuteMatch[1]) * 60 * 1000;
      }
      
      // Seconds
      const secondMatch = value.match(/(\d+)s/);
      if (secondMatch) {
        totalMilliseconds += parseInt(secondMatch[1]) * 1000;
      }
      
      // Milliseconds
      const msMatch = value.match(/(\d+)ms/);
      if (msMatch) {
        totalMilliseconds += parseInt(msMatch[1]);
      }
      
      // Microseconds
      const usMatch = value.match(/(\d+)µs|(\d+)us/);
      if (usMatch) {
        const usValue = usMatch[1] || usMatch[2];
        totalMilliseconds += parseInt(usValue) / 1000;
      }
      
      // If we found time units, return the calculated milliseconds
      if (totalMilliseconds > 0) {
        return totalMilliseconds;
      }
      
      // Otherwise, try to extract numeric values
      const numericMatch = value.replace(/[^0-9.]/g, '');
      if (numericMatch) {
        return parseFloat(numericMatch);
      }
    }
    
    return null;
  };

  // Find the best value for a given operation and metric
  const findBestValue = (operation, metricKey) => {
    const metric = metrics[metricKey];
    if (!metric) return null;
    
    let bestDb = null;
    let bestValue = null;
    
    // Map to store numeric values for comparison
    const dbValues = {};
    
    // First pass: extract and convert all values to comparable numbers
    databases.forEach(db => {
      const dbData = benchmarkData[db];
      if (dbData && dbData[operation]) {
        const value = dbData[operation][metricKey];
        const numericValue = parseValueForComparison(value);
        
        if (numericValue !== null) {
          dbValues[db] = numericValue;
        }
      }
    });
    
    // Second pass: find the best value
    for (const [db, value] of Object.entries(dbValues)) {
      if (bestValue === null) {
        bestValue = value;
        bestDb = db;
      } else if (metric.isHigherBetter) {
        if (value > bestValue) {
          bestValue = value;
          bestDb = db;
        }
      } else {
        if (value < bestValue) {
          bestValue = value;
          bestDb = db;
        }
      }
    }
    
    return bestDb;
  };

  // Get cell style based on whether it's the best value
  const getCellStyle = (db, operation, metricKey) => {
    const bestDb = findBestValue(operation, metricKey);
    if (bestDb === db) {
      return { color: '#4ade80' }; // Green color for best values
    }
    return {};
  };

  // Get value for a given database, operation, and metric
  const getValue = (db, operation, metricKey) => {
    const metric = metrics[metricKey];
    if (!metric) return '-';
    
    const dbData = benchmarkData[db];
    
    if (dbData && dbData[operation] && dbData[operation][metricKey] !== undefined) {
      const value = dbData[operation][metricKey];
      if (value === null || value === '') return '-';
      
      // Try to apply formatting
      try {
        return metric.format(value);
      } catch (e) {
        return value;
      }
    }
    
    return '-';
  };

  // Remove a database from the comparison
  const removeDatabase = (dbToRemove) => {
    setDatabases(databases.filter(db => db !== dbToRemove));
    setBenchmarkData(prevData => {
      const newData = {...prevData};
      delete newData[dbToRemove];
      return newData;
    });
  };

  // Clear all data
  const clearAll = () => {
    setFiles([]);
    setDatabases([]);
    setBenchmarkData({});
  };

  // Group metrics into categories for the tab interface
  const metricGroups = {
    'Performance': ['OPS', 'Total time'],
    'Response Times': ['Mean', 'Max', 'Min'],
    'Percentiles': ['99th', '95th', '75th', '50th', '25th', '1st', 'IQR'],
    'System Resources': ['CPU', 'Memory', 'Reads', 'Writes', 'System load']
  };

  return (
    <div className="flex flex-col min-h-screen bg-gray-900 text-white p-4">
      <h1 className="text-2xl font-bold mb-6">Database Benchmark Dashboard</h1>
      
      {/* File Upload */}
      <div className="mb-6">
        <div className="flex gap-4 mb-2">
          <input
            type="file"
            accept=".csv"
            multiple
            onChange={handleFileUpload}
            ref={fileInputRef}
            className="hidden"
            id="file-upload"
          />
          <label 
            htmlFor="file-upload"
            className="px-4 py-2 bg-blue-600 hover:bg-blue-700 rounded cursor-pointer"
          >
            Upload CSV Files
          </label>
          
          {databases.length > 0 && (
            <button
              onClick={clearAll}
              className="px-4 py-2 bg-red-600 hover:bg-red-700 rounded"
            >
              Clear All
            </button>
          )}
        </div>
        
        {files.length > 0 && (
          <div className="text-sm text-gray-300 mb-4">
            Loaded {files.length} file(s): {files.map(f => f.name).join(', ')}
          </div>
        )}
      </div>
      
      {/* Tab Navigation - First level: Metric Groups */}
      {databases.length > 0 && (
        <div className="mb-4">
          <div className="flex flex-wrap gap-2 mb-4">
            {Object.entries(metricGroups).map(([groupName, groupMetrics]) => (
              <div key={groupName} className="flex flex-col">
                <h3 className="text-gray-400 text-sm mb-1">{groupName}</h3>
                <div className="flex flex-wrap gap-1">
                  {groupMetrics.map(metricKey => (
                    <button
                      key={metricKey}
                      className={`px-3 py-1 text-sm rounded ${
                        activeTab === metricKey 
                          ? 'bg-blue-600 text-white' 
                          : 'bg-gray-800 text-gray-300 hover:bg-gray-700'
                      }`}
                      onClick={() => setActiveTab(metricKey)}
                    >
                      {metrics[metricKey].title}
                    </button>
                  ))}
                </div>
              </div>
            ))}
          </div>
        </div>
      )}
      
      {/* Metric Description */}
      {databases.length > 0 && metrics[activeTab] && (
        <div className="text-sm text-gray-400 mb-4">
          {metrics[activeTab].description}
        </div>
      )}
      
      {/* Benchmark Table */}
      {databases.length > 0 ? (
        <div className="overflow-x-auto">
          <table className="w-full border-collapse">
            <thead>
              <tr className="border-b border-gray-700">
                <th className="text-left p-2">Benchmark</th>
                {databases.map(db => (
                  <th key={db} className="text-left p-2">
                    <div className="flex items-center justify-between">
                      <span>{db}</span>
                      <button 
                        onClick={() => removeDatabase(db)}
                        className="ml-2 text-gray-500 hover:text-red-500"
                        title="Remove database"
                      >
                        ✕
                      </button>
                    </div>
                  </th>
                ))}
              </tr>
            </thead>
            <tbody>
              {operations.map(operation => (
                <tr key={operation} className="border-b border-gray-800 hover:bg-gray-800">
                  <td className="p-2">{operation}</td>
                  {databases.map(db => (
                    <td 
                      key={`${db}-${operation}`} 
                      className="p-2"
                      style={getCellStyle(db, operation, activeTab)}
                    >
                      {getValue(db, operation, activeTab)}
                    </td>
                  ))}
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      ) : (
        <div className="text-center text-gray-400 py-12">
          <p>Upload CSV files to compare database benchmarks</p>
          <p className="text-sm mt-2">Files should contain benchmark results with these columns:</p>
          <p className="text-xs mt-1 text-gray-500">
            Test, Total time, Mean, Max, 99th, 95th, 75th, 50th, 25th, 1st, Min, IQR, OPS, CPU, Memory, Reads, Writes, System load
          </p>
        </div>
      )}
    </div>
  );
}

export default App;