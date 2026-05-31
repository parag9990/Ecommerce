import type {
  FilterOption,
  SegmentFilters
} from "../types";
import {
  channelOptions,
  deviceTypeOptions,
  sourceOptions,
  userTypeOptions
} from "../types";

type SegmentFilterPanelProps = {
  value: SegmentFilters;
  onChange: (value: SegmentFilters) => void;
};

export function SegmentFilterPanel({ value, onChange }: SegmentFilterPanelProps) {
  return (
    <div className="grid gap-3 md:grid-cols-2 xl:grid-cols-4">
      <FilterSelect
        id="device-filter"
        label="Device"
        options={deviceTypeOptions}
        value={value.deviceType}
        onChange={(deviceType) => onChange({ ...value, deviceType })}
      />
      <FilterSelect
        id="channel-filter"
        label="Channel"
        options={channelOptions}
        value={value.channel}
        onChange={(channel) => onChange({ ...value, channel })}
      />
      <FilterSelect
        id="source-filter"
        label="Source"
        options={sourceOptions}
        value={value.source}
        onChange={(source) => onChange({ ...value, source })}
      />
      <FilterSelect
        id="user-type-filter"
        label="User Type"
        options={userTypeOptions}
        value={value.userType}
        onChange={(userType) => onChange({ ...value, userType })}
      />
    </div>
  );
}

type FilterSelectProps<T extends string> = {
  id: string;
  label: string;
  options: ReadonlyArray<FilterOption<T>>;
  value: T;
  onChange: (value: T) => void;
};

function FilterSelect<T extends string>({
  id,
  label,
  options,
  value,
  onChange
}: FilterSelectProps<T>) {
  return (
    <label className="block" htmlFor={id}>
      <span className="mb-1 block text-xs font-semibold uppercase text-zinc-500">
        {label}
      </span>
      <select
        className="h-10 w-full rounded-md border border-zinc-300 bg-white px-3 text-sm capitalize outline-none transition-colors focus:border-zinc-950 focus:ring-2 focus:ring-zinc-950/10"
        id={id}
        onChange={(event) => onChange(event.target.value as T)}
        value={value}
      >
        {options.map((option) => (
          <option key={option.value} value={option.value}>
            {option.label}
          </option>
        ))}
      </select>
    </label>
  );
}
