'use client';

import { Crosshair, LocateFixed, MapPin, Minus, Plus } from 'lucide-react';
import { useCallback, useEffect, useRef, useState } from 'react';
import type { CircleMarker, Map as LeafletMap } from 'leaflet';
import { Action, FieldInput } from './components';

type Point = [number, number];
type ViewRequest = { point: Point };

function parsePoint(value: string): Point | null {
  const parts = value.split(',').map((part) => part.trim());
  if (parts.length !== 2 || parts.some((part) => !part)) return null;
  const [lat, lng] = parts.map(Number);
  return Math.abs(lat) <= 90 && Math.abs(lng) <= 180 ? [lat, lng] : null;
}

function LocationMap({
  coordinates,
  view = null,
  onPick,
}: {
  coordinates: string;
  view?: ViewRequest | null;
  onPick?: (point: Point) => void;
}) {
  const element = useRef<HTMLDivElement>(null);
  const firstPoint = useRef(parsePoint(coordinates));
  const [layers, setLayers] = useState<{
    map: LeafletMap;
    marker: CircleMarker;
  }>();
  const [notice, setNotice] = useState('');

  useEffect(() => {
    let mounted = true;
    let map: LeafletMap | undefined;
    let observer: ResizeObserver | undefined;
    import('leaflet')
      .then((leaflet) => {
        if (!mounted || !element.current) return;
        map = leaflet
          .map(element.current, {
            zoomControl: false,
            scrollWheelZoom: false,
            minZoom: 3,
            maxZoom: 19,
          })
          .setView(
            firstPoint.current ?? [16.2, 107.8],
            firstPoint.current ? 16 : 5,
          );
        map.attributionControl.setPrefix(false);
        const tiles = leaflet.tileLayer(
          'https://tile.openstreetmap.org/{z}/{x}/{y}.png',
          {
            attribution:
              '© <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a>',
            maxZoom: 19,
            noWrap: true,
          },
        );
        tiles.on('tileerror', () =>
          setNotice('Một phần nền bản đồ chưa tải được.'),
        );
        tiles.on('load', () => {
          if (element.current?.querySelector('.leaflet-tile-loaded'))
            setNotice('');
        });
        tiles.addTo(map);
        if (onPick) {
          map.on('click', (event) => {
            const point = event.latlng.wrap();
            onPick([point.lat, point.lng]);
          });
        }
        const marker = leaflet.circleMarker([0, 0], {
          radius: 8,
          weight: 3,
          color: '#fff',
          fillColor: '#16453a',
          fillOpacity: 1,
          interactive: false,
        });
        observer = new ResizeObserver(() =>
          map?.invalidateSize({ pan: false }),
        );
        observer.observe(element.current);
        setLayers({ map, marker });
      })
      .catch(() => {
        if (mounted)
          setNotice(
            'Bản đồ chưa tải được. Bạn có thể dùng vị trí thiết bị hoặc nhập tọa độ.',
          );
      });
    return () => {
      mounted = false;
      observer?.disconnect();
      map?.remove();
    };
  }, [onPick]);

  useEffect(() => {
    if (!layers) return;
    const point = parsePoint(coordinates);
    if (point) layers.marker.setLatLng(point).addTo(layers.map);
    else layers.marker.remove();
  }, [coordinates, layers]);

  useEffect(() => {
    if (layers && view) layers.map.setView(view.point, 16);
  }, [layers, view]);

  function pinCenter() {
    if (!layers || !onPick) return;
    const point = layers.map.getCenter().wrap();
    onPick([point.lat, point.lng]);
  }

  return (
    <div className="site-map">
      <div className="site-map-tools">
        {onPick && (
          <Action
            type="button"
            className="secondary"
            disabled={!layers}
            onClick={pinCenter}
            leftSection={<Crosshair size={17} aria-hidden />}
          >
            Ghim tâm
          </Action>
        )}
        <Action
          type="button"
          className="secondary site-map-zoom"
          aria-label="Phóng to bản đồ"
          disabled={!layers}
          onClick={() => layers?.map.zoomIn()}
        >
          <Plus size={18} aria-hidden />
        </Action>
        <Action
          type="button"
          className="secondary site-map-zoom"
          aria-label="Thu nhỏ bản đồ"
          disabled={!layers}
          onClick={() => layers?.map.zoomOut()}
        >
          <Minus size={18} aria-hidden />
        </Action>
      </div>
      <div
        className="site-map-canvas"
        role="region"
        aria-label={
          onPick ? 'Bản đồ chọn vị trí dãy trọ' : 'Bản đồ vị trí đã lưu'
        }
        data-ready={!!layers}
        ref={element}
      />
      {notice && (
        <p role="status" className="subtle">
          {notice}
        </p>
      )}
    </div>
  );
}

export function SavedSiteMap({ coordinates }: { coordinates: string }) {
  return <LocationMap coordinates={coordinates} />;
}

export function SiteLocationField({
  initialCoordinates = '',
  expanded = false,
}: {
  initialCoordinates?: string;
  expanded?: boolean;
}) {
  const [open, setOpen] = useState(false);
  const [coordinates, setCoordinates] = useState(initialCoordinates);
  const [view, setView] = useState<ViewRequest | null>(null);
  const [locating, setLocating] = useState(false);
  const [message, setMessage] = useState('');
  const request = useRef(0);

  const pick = useCallback(([lat, lng]: Point) => {
    request.current++;
    setLocating(false);
    setCoordinates(`${lat.toFixed(6)}, ${lng.toFixed(6)}`);
    setMessage('');
  }, []);

  function locate() {
    if (!navigator.geolocation || !window.isSecureContext) {
      setMessage(
        'Trình duyệt cần HTTPS hoặc localhost để lấy vị trí. Bạn vẫn có thể ghim trên bản đồ.',
      );
      return;
    }
    const id = ++request.current;
    setLocating(true);
    setMessage('Đang lấy vị trí thiết bị…');
    navigator.geolocation.getCurrentPosition(
      ({ coords }) => {
        if (request.current !== id) return;
        const point: Point = [coords.latitude, coords.longitude];
        pick(point);
        setView({ point });
        setMessage('Đã lấy vị trí thiết bị. Bạn có thể ghim lại trên bản đồ.');
      },
      (error) => {
        if (request.current !== id) return;
        setLocating(false);
        setMessage(
          error.code === 1
            ? 'Trình duyệt chưa cấp quyền vị trí. Hãy bấm trên bản đồ để ghim.'
            : 'Chưa lấy được vị trí thiết bị. Hãy bấm trên bản đồ để ghim.',
        );
      },
      { enableHighAccuracy: true, timeout: 10000, maximumAge: 0 },
    );
  }

  const content = (
    <div className="form">
      <Action
        type="button"
        className="secondary"
        disabled={locating}
        onClick={locate}
        leftSection={<LocateFixed size={18} aria-hidden />}
      >
        Dùng vị trí hiện tại
      </Action>
      <p className="subtle">Bấm vào bản đồ để ghim điểm muốn lưu.</p>
      {(expanded || open) && (
        <LocationMap coordinates={coordinates} view={view} onPick={pick} />
      )}
      {message && (
        <p role="status" className="subtle">
          {message}
        </p>
      )}
      {parsePoint(coordinates) && (
        <div className="room-toolbar">
          <span role="status">Vị trí đã chọn: {coordinates}</span>
          <Action
            type="button"
            className="secondary"
            onClick={() => {
              request.current++;
              setLocating(false);
              setCoordinates('');
              setMessage('');
              setView(null);
            }}
          >
            Bỏ ghim
          </Action>
        </div>
      )}
      <details>
        <summary>Nhập tọa độ thủ công</summary>
        <label>
          Tọa độ
          <FieldInput
            name="coordinates"
            value={coordinates}
            maxLength={65}
            placeholder="Ví dụ 10.7769, 106.7009"
            onChange={(event) => {
              request.current++;
              setLocating(false);
              setCoordinates(event.currentTarget.value);
            }}
            onBlur={(event) => {
              const point = parsePoint(event.currentTarget.value);
              if (point) setView({ point });
            }}
          />
        </label>
      </details>
    </div>
  );

  if (expanded) return content;

  return (
    <details
      className="site-location-picker"
      onToggle={(event) => {
        setOpen(event.currentTarget.open);
        if (!event.currentTarget.open) {
          request.current++;
          setLocating(false);
          setMessage('');
          setView(null);
        }
      }}
    >
      <summary>
        <MapPin size={18} aria-hidden /> Chọn vị trí trên bản đồ
      </summary>
      {content}
    </details>
  );
}
