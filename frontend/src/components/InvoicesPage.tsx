import { useState, useEffect, useCallback } from 'react';
import {
  useReactTable,
  getCoreRowModel,
  getSortedRowModel,
  flexRender,
  createColumnHelper,
  type SortingState,
} from '@tanstack/react-table';

interface Employee {
  id: number;
  name: string;
}

interface Invoice {
  id: number;
  date: string;
  employee_id: number;
  employee_name: string;
  total: number;
}

interface ApiResponse<T> {
  [key: string]: T[];
}

const columnHelper = createColumnHelper<Invoice>();

const columns = [
  columnHelper.accessor('id', {
    header: 'ID',
    cell: (info) => info.getValue(),
  }),
  columnHelper.accessor('date', {
    header: 'Date',
    cell: (info) => info.getValue(),
  }),
  columnHelper.accessor('employee_id', {
    header: 'Employee ID',
    cell: (info) => info.getValue(),
  }),
  columnHelper.accessor('employee_name', {
    header: 'Employee Name',
    cell: (info) => info.getValue(),
  }),
  columnHelper.accessor('total', {
    header: 'Total',
    cell: (info) => `$${info.getValue().toFixed(2)}`,
  }),
];

function InvoicesPage() {
  const [employees, setEmployees] = useState<Employee[]>([]);
  const [invoices, setInvoices] = useState<Invoice[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [sorting, setSorting] = useState<SortingState>([]);

  // Calculate previous month dates
  const getPreviousMonthDates = () => {
    const now = new Date();
    const prevMonth = new Date(now.getFullYear(), now.getMonth() - 1, 1);
    const startDate = prevMonth.toISOString().split('T')[0];
    const endDate = new Date(prevMonth.getFullYear(), prevMonth.getMonth() + 1, 0).toISOString().split('T')[0];
    return { startDate, endDate };
  };

  const { startDate: defaultStart, endDate: defaultEnd } = getPreviousMonthDates();

  const [startDate, setStartDate] = useState(defaultStart);
  const [endDate, setEndDate] = useState(defaultEnd);
  const [selectedEmployeeId, setSelectedEmployeeId] = useState<string>('');

  // Fetch employees on mount
  useEffect(() => {
    fetch('/api/employees')
      .then((res) => res.json())
      .then((data: ApiResponse<Employee>) => setEmployees(data.employees))
      .catch((err) => console.error('Failed to fetch employees:', err));
  }, []);

  const fetchInvoices = useCallback(async () => {
    setLoading(true);
    setError(null);

    const params = new URLSearchParams({
      start_date: startDate,
      end_date: endDate,
    });

    if (selectedEmployeeId) {
      params.append('employee_id', selectedEmployeeId);
    }

    try {
      const res = await fetch(`/api/invoices?${params}`);
      const data = await res.json();

      if (!res.ok) {
        setError(data.error || 'Failed to fetch invoices');
        setInvoices([]);
      } else {
        setInvoices(data.invoices);
      }
    } catch {
      setError('Network error occurred');
      setInvoices([]);
    } finally {
      setLoading(false);
    }
  }, [startDate, endDate, selectedEmployeeId]);

  // Debounced invoice fetch
  useEffect(() => {
    const timeoutId = setTimeout(() => {
      fetchInvoices();
    }, 300);
    return () => clearTimeout(timeoutId);
  }, [fetchInvoices]);

  const table = useReactTable({
    data: invoices,
    columns,
    state: {
      sorting,
    },
    onSortingChange: setSorting,
    getCoreRowModel: getCoreRowModel(),
    getSortedRowModel: getSortedRowModel(),
  });

  return (
    <div className="min-h-screen bg-gray-50 p-6">
      <div className="max-w-7xl mx-auto">
        <h1 className="text-3xl font-bold text-gray-900 mb-8">Invoices</h1>

        {/* Form Section */}
        <div className="bg-white p-6 rounded-lg shadow-sm mb-6">
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            <div>
              <label htmlFor="start-date" className="block text-sm font-medium text-gray-700 mb-1">
                Start Date
              </label>
              <input
                id="start-date"
                type="date"
                value={startDate}
                onChange={(e) => setStartDate(e.target.value)}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              />
            </div>

            <div>
              <label htmlFor="end-date" className="block text-sm font-medium text-gray-700 mb-1">
                End Date
              </label>
              <input
                id="end-date"
                type="date"
                value={endDate}
                onChange={(e) => setEndDate(e.target.value)}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              />
            </div>

            <div>
              <label htmlFor="employee" className="block text-sm font-medium text-gray-700 mb-1">
                Employee
              </label>
              <select
                id="employee"
                value={selectedEmployeeId}
                onChange={(e) => setSelectedEmployeeId(e.target.value)}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              >
                <option value="">All Employees</option>
                {employees.map((emp) => (
                  <option key={emp.id} value={emp.id.toString()}>
                    {emp.name}
                  </option>
                ))}
              </select>
            </div>
          </div>
        </div>

        {/* Error Message */}
        {error && (
          <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded mb-6">
            {error}
          </div>
        )}

        {/* Loading or Table */}
        {loading ? (
          <div className="flex justify-center items-center py-12">
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
          </div>
        ) : (
          <div className="bg-white rounded-lg shadow-sm overflow-hidden">
            <div className="overflow-x-auto">
              <table className="min-w-full divide-y divide-gray-200">
                <thead className="bg-gray-50">
                  {table.getHeaderGroups().map((headerGroup) => (
                    <tr key={headerGroup.id}>
                      {headerGroup.headers.map((header) => (
                        <th
                          key={header.id}
                          className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider cursor-pointer hover:bg-gray-100"
                          onClick={header.column.getToggleSortingHandler()}
                        >
                          {header.isPlaceholder
                            ? null
                            : flexRender(header.column.columnDef.header, header.getContext())}
                          {{
                            asc: ' 🔼',
                            desc: ' 🔽',
                          }[header.column.getIsSorted() as string] ?? null}
                        </th>
                      ))}
                    </tr>
                  ))}
                </thead>
                <tbody className="bg-white divide-y divide-gray-200">
                  {table.getRowModel().rows.map((row) => (
                    <tr key={row.id} className="hover:bg-gray-50">
                      {row.getVisibleCells().map((cell) => (
                        <td key={cell.id} className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                          {flexRender(cell.column.columnDef.cell, cell.getContext())}
                        </td>
                      ))}
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
            {invoices.length === 0 && !loading && (
              <div className="text-center py-12 text-gray-500">
                No invoices found for the selected criteria.
              </div>
            )}
          </div>
        )}
      </div>
    </div>
  );
}

export default InvoicesPage;