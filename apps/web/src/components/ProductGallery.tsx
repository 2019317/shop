'use client'

import { useState } from 'react'

export default function ProductGallery({
  images,
  title,
}: {
  images: { url: string; alt: string }[]
  title: string
}) {
  const [active, setActive] = useState(0)
  const list = images.length ? images : [{ url: '', alt: title }]

  return (
    <div>
      {/* eslint-disable-next-line @next/next/no-img-element */}
      <img className="gallery-main" src={list[active].url} alt={list[active].alt || title} />
      {list.length > 1 && (
        <div className="gallery-thumbs">
          {list.map((img, i) => (
            // eslint-disable-next-line @next/next/no-img-element
            <img
              key={i}
              src={img.url}
              alt={img.alt || title}
              className={i === active ? 'active' : ''}
              onClick={() => setActive(i)}
            />
          ))}
        </div>
      )}
    </div>
  )
}
