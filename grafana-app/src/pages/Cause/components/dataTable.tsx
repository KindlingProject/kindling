import React from 'react';
import { useTable } from 'react-table';
import { css, cx } from '@emotion/css';

interface TableProps {
  data: any[],
  columns: any[]
}
const DataTable = ({ data, columns}: TableProps) => {
  const {
    getTableProps,
    getTableBodyProps,
    headerGroups,
    rows,
    prepareRow,
  } = useTable({
    columns,
    data,
  });

  return (
    <table {...getTableProps()} className={cx(custom_table)}>
      <thead>
        {headerGroups.map((headerGroup, idx) => (
          <tr {...headerGroup.getHeaderGroupProps()} key={`header_tr_${idx}`}>
            {headerGroup.headers.map((column, idx) => (
              <th {...column.getHeaderProps()} key={`header_th_${idx}`}>{column.render('title')}</th>
            ))}
          </tr>
        ))}
      </thead>
      <tbody {...getTableBodyProps()}>
        {rows.map((row, idx) => {
          prepareRow(row);
          return (
            <tr {...row.getRowProps()} key={`body_tr_${idx}`}>
              {row.cells.map((cell, idx) => {
                return (
                  <td {...cell.getCellProps()} key={`body_td_${idx}`}>
                    {cell.render('Cell')}
                  </td>
                );
              })}
            </tr>
          );
        })}
      </tbody>
    </table>
  );
};

const custom_table = css`
  width: 100%;
  margin: 10px 0;
  tr {
    height: max-content;
    border-bottom: 1px solid rgba(204, 204, 220, 0.12);
    display: flex;
    th {
      min-width: 80px;
      display: flex;
      justify-content: flex-start;
      flex: 1;
      align-items: center;
      padding: 0 6px;
      height: 35px;
    }
    td {
      min-width: 80px;
      display: flex;
      justify-content: flex-start;
      align-items: center;
      flex: 1;
      padding: 6px;
      height: initial;
      white-space: pre-wrap;
      border-right: 1px solid rgba(204, 204, 220, 0.12);
      &:last-child {
        border-right: none;
      }
    }
  }
`;
export default DataTable;
