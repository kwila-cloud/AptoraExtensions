import { useState, useEffect, useCallback } from "react";
import { useSearchParams } from "react-router-dom";
import {
  useReactTable,
  getCoreRowModel,
  getSortedRowModel,
  flexRender,
  createColumnHelper,
  type SortingState,
} from "@tanstack/react-table";

interface Employee {
  id: number;
  name: string;
}

interface Invoice {
  number: number;
  date: string;
  employee_name: string;
  subtotal: number;
}

interface ApiResponse<T> {
  [key: string]: T[];
}

const columnHelper = createColumnHelper<Invoice>();

const columns = [
  columnHelper.accessor("number", {
    header: "Number",
    cell: (info) => info.getValue(),
  }),
  columnHelper.accessor("date", {
    header: "Date",
    cell: (info) => info.getValue(),
  }),
  columnHelper.accessor("employee_name", {
    header: "Employee Name",
    cell: (info) => info.getValue(),
  }),
  columnHelper.accessor("subtotal", {
    header: "Subtotal",
    cell: (info) => `$${info.getValue().toFixed(2)}`,
  }),
];

function InvoicesPage() {
  const [searchParams, setSearchParams] = useSearchParams();
  const [employees, setEmployees] = useState<Employee[]>([]);
  const [invoices, setInvoices] = useState<Invoice[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [sorting, setSorting] = useState<SortingState>([
    { id: "date", desc: false },
  ]);

  // Calculate previous month dates
  const getPreviousMonthDates = () => {
    const now = new Date();
    const prevMonth = new Date(now.getFullYear(), now.getMonth() - 1, 1);
    const startDate = prevMonth.toISOString().split("T")[0];
    const endDate = new Date(
      prevMonth.getFullYear(),
      prevMonth.getMonth() + 1,
      0,
    )
      .toISOString()
      .split("T")[0];
    return { startDate, endDate };
  };

  const { startDate: defaultStart, endDate: defaultEnd } =
    getPreviousMonthDates();

  // Initialize state from URL params or defaults
  const [startDate, setStartDate] = useState(
    searchParams.get("start_date") || defaultStart,
  );
  const [endDate, setEndDate] = useState(
    searchParams.get("end_date") || defaultEnd,
  );
  const [selectedEmployee, setSelectedEmployee] = useState<string>(
    searchParams.get("employee") || "",
  );
  const [filtersExpanded, setFiltersExpanded] = useState(true);

  // Update URL params when filters change
  useEffect(() => {
    const params = new URLSearchParams();
    params.set("start_date", startDate);
    params.set("end_date", endDate);
    if (selectedEmployee) {
      params.set("employee", selectedEmployee);
    }
    setSearchParams(params, { replace: true });
  }, [startDate, endDate, selectedEmployee, setSearchParams]);

  // Fetch employees on mount and sort alphabetically
  useEffect(() => {
    const controller = new AbortController();

    const fetchEmployees = async () => {
      try {
        const res = await fetch("/api/employees", {
          signal: controller.signal,
        });

        if (!res.ok) {
          const errorData = await res
            .json()
            .catch(() => ({ error: "Failed to fetch employees" }));
          console.error(
            "Failed to fetch employees:",
            errorData.error || "Unknown error",
          );
          return;
        }

        const data: ApiResponse<Employee> = await res.json();

        if (data.employees && Array.isArray(data.employees)) {
          const sorted = [...data.employees].sort((a, b) =>
            a.name.localeCompare(b.name),
          );
          setEmployees(sorted);
        }
      } catch (err) {
        // Ignore AbortError on unmount
        if (err instanceof Error && err.name !== "AbortError") {
          console.error("Failed to fetch employees:", err.message);
        }
      }
    };

    fetchEmployees();

    return () => {
      controller.abort();
    };
  }, []);

  const fetchInvoices = useCallback(async () => {
    setLoading(true);
    setError(null);

    const params = new URLSearchParams({
      start_date: startDate,
      end_date: endDate,
    });

    if (selectedEmployee) {
      params.append("employee", selectedEmployee);
    }

    try {
      const res = await fetch(`/api/invoices?${params}`);
      const data = await res.json();

      if (!res.ok) {
        setError(data.error || "Failed to fetch invoices");
        setInvoices([]);
      } else {
        setInvoices(data.invoices);
      }
    } catch {
      setError("Network error occurred");
      setInvoices([]);
    } finally {
      setLoading(false);
    }
  }, [startDate, endDate, selectedEmployee]);

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
    <div className="h-screen bg-gray-50 p-3 md:p-6 flex flex-col overflow-hidden">
      <div className="max-w-7xl mx-auto w-full flex flex-col h-full">
        <h1 className="text-2xl md:text-3xl font-bold text-gray-900 mb-4 md:mb-8 flex-shrink-0">
          Invoices
        </h1>

        {/* Form Section */}
        <div className="bg-white rounded-lg shadow-sm mb-3 md:mb-6 flex-shrink-0">
          {/* Mobile Toggle Button */}
          <button
            onClick={() => setFiltersExpanded(!filtersExpanded)}
            className="md:hidden w-full px-4 py-3 flex items-center justify-between text-left font-medium text-gray-900"
          >
            <span>Filters</span>
            <span className="text-gray-500">{filtersExpanded ? "▲" : "▼"}</span>
          </button>

          {/* Filters */}
          <div
            className={`${filtersExpanded ? "block" : "hidden"} md:block p-4 md:p-6`}
          >
            <div className="grid grid-cols-1 md:grid-cols-3 gap-3 md:gap-4">
              <div>
                <label
                  htmlFor="start-date"
                  className="block text-sm font-medium text-gray-700 mb-1"
                >
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
                <label
                  htmlFor="end-date"
                  className="block text-sm font-medium text-gray-700 mb-1"
                >
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
                <label
                  htmlFor="employee"
                  className="block text-sm font-medium text-gray-700 mb-1"
                >
                  Employee
                </label>
                <select
                  id="employee"
                  value={selectedEmployee}
                  onChange={(e) => setSelectedEmployee(e.target.value)}
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                >
                  <option value="">All Employees</option>
                  {employees.map((emp) => (
                    <option key={emp.id} value={emp.name}>
                      {emp.name}
                    </option>
                  ))}
                </select>
              </div>
            </div>
          </div>
        </div>

        {/* Error Message */}
        {error && (
          <div className="bg-red-50 border border-red-200 text-red-700 px-3 md:px-4 py-2 md:py-3 rounded mb-3 md:mb-6 flex-shrink-0 text-sm">
            {error}
          </div>
        )}

        {/* Loading or Table */}
        {loading ? (
          <div className="flex justify-center items-center py-12 flex-1">
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
          </div>
        ) : (
          <div className="bg-white rounded-lg shadow-sm overflow-hidden flex-1 flex flex-col min-h-0">
            {/* Invoice Count */}
            {invoices.length > 0 && (
              <div className="px-3 md:px-6 py-2 md:py-3 border-b border-gray-200 bg-gray-50 flex-shrink-0">
                <p className="text-xs md:text-sm text-gray-700">
                  Showing{" "}
                  <span className="font-semibold">{invoices.length}</span>{" "}
                  invoice{invoices.length !== 1 ? "s" : ""}
                </p>
              </div>
            )}
            <div className="overflow-auto flex-1">
              <table className="min-w-full">
                <thead className="sticky top-0 z-10">
                  {table.getHeaderGroups().map((headerGroup) => (
                    <tr
                      key={headerGroup.id}
                      className="shadow-[0_2px_0_0_rgba(0,0,0,0.1)]"
                    >
                      {headerGroup.headers.map((header) => (
                        <th
                          key={header.id}
                          className="px-3 md:px-6 py-2 md:py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider cursor-pointer hover:bg-gray-100 bg-gray-50"
                          onClick={header.column.getToggleSortingHandler()}
                        >
                          <div className="flex items-center gap-1">
                            {header.isPlaceholder
                              ? null
                              : flexRender(
                                  header.column.columnDef.header,
                                  header.getContext(),
                                )}
                            <span className="text-gray-400">
                              {{
                                asc: "▲",
                                desc: "▼",
                              }[header.column.getIsSorted() as string] ?? ""}
                            </span>
                          </div>
                        </th>
                      ))}
                    </tr>
                  ))}
                </thead>
                <tbody className="bg-white divide-y divide-gray-200">
                  {table.getRowModel().rows.map((row) => (
                    <tr key={row.id} className="hover:bg-gray-50">
                      {row.getVisibleCells().map((cell) => (
                        <td
                          key={cell.id}
                          className="px-3 md:px-6 py-2 md:py-4 whitespace-nowrap text-xs md:text-sm text-gray-900"
                        >
                          {flexRender(
                            cell.column.columnDef.cell,
                            cell.getContext(),
                          )}
                        </td>
                      ))}
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
            {invoices.length === 0 && !loading && (
              <div className="text-center py-8 md:py-12 text-sm md:text-base text-gray-500">
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
