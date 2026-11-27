'use client';

import { useState } from 'react';
import { FieldSelect, FieldInput } from './components';
import type { PageData } from './view';

export function RoomPricing({
  room,
  policies,
}: {
  room: PageData['RoomDetail'];
  policies: PageData['Policies'];
}) {
  const canChoosePolicy = !!policies?.length;
  const [selected, setSelected] = useState(room?.PolicyID || '');
  const usePolicy = canChoosePolicy && !!selected;
  return (
    <>
      {canChoosePolicy && (
        <label>
          Mẫu giá và phí
          <FieldSelect
            name="policy"
            value={selected}
            onChange={(event) => setSelected(event.currentTarget.value)}
          >
            <option value="">Nhập giá riêng cho phòng</option>
            {policies.map((policy) => (
              <option key={policy.ID} value={policy.ID}>
                {policy.Name}
              </option>
            ))}
          </FieldSelect>
        </label>
      )}
      <div className="form" hidden={usePolicy}>
        <label>
          Tiền thuê mỗi tháng (đồng)
          <FieldInput
            name="rent"
            type="number"
            inputMode="numeric"
            min={0}
            max={100000000}
            required={!canChoosePolicy}
            defaultValue={room?.Rent}
          />
        </label>
        <div className="pair">
          <label>
            Giá điện (đồng/kWh)
            <FieldInput
              name="electricRate"
              type="number"
              inputMode="numeric"
              min={0}
              max={100000}
              required={!canChoosePolicy}
              defaultValue={room?.ElectricRate}
            />
          </label>
          <label>
            Giá nước (đồng/m³)
            <FieldInput
              name="waterRate"
              type="number"
              inputMode="numeric"
              min={0}
              max={1000000}
              required={!canChoosePolicy}
              defaultValue={room?.WaterRate}
            />
          </label>
        </div>
      </div>
    </>
  );
}
